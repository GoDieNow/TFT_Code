package dbManager

import (
	"errors"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/remeh/sizedwaitgroup"
	"gitlab.com/cyclops-utilities/datamodels"
	"github.com/GoDieNow/TFT_Code/services/eventsengine/models"
	l "gitlab.com/cyclops-utilities/logging"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	statusDuplicated = iota
	statusFail
	statusMissing
	statusOK

	innerIncluded
	lowerIncluded
	notIncluded
	outterIncluded
	upperIncluded

	stateActive     = "active"
	stateError      = "error"
	stateInactive   = "inactive"
	stateSuspended  = "suspended"
	stateTerminated = "terminated"

	bigBang   = int64(0)
	endOfTime = int64(32503680000)
)

// DbParameter is the struct defined to group and contain all the methods
// that interact with the database.
// On it there is the following parameters:
// - connStr: strings with the connection information to the database
// - Db: a gorm.DB pointer to the db to invoke all the db methods
type DbParameter struct {
	connStr string
	Db      *gorm.DB
	Metrics map[string]*prometheus.GaugeVec
}

var (
	eeTime  float64
	eeTotal float64
)

// New is the function to create the struct DbParameter.
// Parameters:
// - dbConn: strings with the connection information to the database
// - tables: array of interfaces that will contains the models to migrate
// to the database on initialization
// Returns:
// - DbParameter: struct to interact with dbManager functionalities
func New(dbConn string, tables ...interface{}) *DbParameter {

	l.Trace.Printf("[DB] Gerenating new DBParameter.\n")

	var (
		dp  DbParameter
		err error
	)

	dp.connStr = dbConn

	dp.Db, err = gorm.Open(postgres.Open(dbConn), &gorm.Config{})

	if err != nil {

		l.Error.Printf("[DB] Error opening connection. Error: %v\n", err)

	}

	l.Trace.Printf("[DB] Migrating tables.\n")

	//Database migration, it handles everything
	dp.Db.AutoMigrate(tables...)

	l.Trace.Printf("[DB] Generating hypertables.\n")

	// Hypertables creation for timescaledb in case of needed
	//dp.Db.Exec("SELECT create_hypertable('" + dp.Db.NewScope(&models.TABLE).TableName() + "', 'TIMESCALE-ROW-INDEX');")

	return &dp

}

// AddEvent job is to register the new event on its corresponding resource state.
// Parameters:
// - event: a reference containing the new event in the resource.
// Returns:
// - e: error raised in case of problems
func (d *DbParameter) AddEvent(event models.Event) (e error) {

	l.Trace.Printf("[DB] Attempting to register a new in the resource [ %v ] from the account [ %v ].\n", event.ResourceID, event.Account)

	now := time.Now().UnixNano()

	var states []*models.State
	var s, state models.State

	patternSearch := models.State{
		Account:    event.Account,
		ResourceID: event.ResourceID,
		Region:     event.Region,
	}

	patternUpdate := models.State{
		TimeFrom:  event.TimeFrom,
		EventTime: event.EventTime,
		LastEvent: func(s string) *string { return &s }(stateTerminated),
	}

	s = patternSearch
	s.MetaData = event.MetaData

	// First check is there's anything from same resourceid with different metadata,
	// if so, terminate them all
	if e = d.Db.Where(patternSearch).Not(models.State{
		MetaData: event.MetaData,
	}).Not(models.State{
		LastEvent: func(s string) *string { return &s }(stateTerminated),
	}).Find(&states).Error; e != nil {

		l.Warning.Printf("[DB] Something went wrong while checking if the resource already has registered states in the system. Error: %v\n", e)

	}

	if len(states) != 0 {

		l.Debug.Printf("[DB] We found [ %v ] states linked to the resource [ %v ] proceeding to mark them as terminated...\n", len(states), event.ResourceID)

		if e = d.UpdateStates(states, patternUpdate); e != nil {

			l.Warning.Printf("[DB] Something went wrong while terminating the other states linked to the resource [ %v ]. Error: %v\n", event.ResourceID, e)

		} else {

			d.Metrics["count"].With(prometheus.Labels{"type": "States updated"}).Add(float64(len(states)))

		}

	}

	// check if Status is the same
	// if so, update the last time event field
	// if diferent, copy actual state to history, then change state into new values from event
	if r := d.Db.Where(&s).First(&state).Error; errors.Is(r, gorm.ErrRecordNotFound) {

		l.Debug.Printf("[DB] The resource [ %v ] doesn't have a linked state in the system, proceeding to generate one...\n", event.ResourceID)

		s.ResourceName = event.ResourceName
		s.TimeTo = endOfTime
		s.TimeFrom = *event.EventTime
		s.ResourceType = event.ResourceType
		s.LastEvent = event.LastEvent
		s.EventTime = event.EventTime

		if e = d.Db.Create(&s).Error; e != nil {

			l.Warning.Printf("[DB] Something went wrong while trying to create the state. Error: %v\n", e)

		} else {

			d.Metrics["count"].With(prometheus.Labels{"type": "States created"}).Inc()

		}

	} else {

		l.Debug.Printf("[DB] The resource [ %v ] seems to be already in the system with state [ %v ], proceeding to update its state...\n", event.ResourceID, *state.LastEvent)

		if *state.LastEvent == *event.LastEvent {

			if e = d.Db.Model(&state).Updates(models.State{
				EventTime: event.EventTime,
			}).Error; e != nil {

				l.Warning.Printf("[DB] Something went wrong while updating the state. Error: %v\n", e)

			} else {

				d.Metrics["count"].With(prometheus.Labels{"type": "States updated"}).Inc()

			}

		} else {

			l.Debug.Printf("[DB] The resource [ %v ] seems to have a previous different state: [ %v ] -> [ %v ], proceeding to update it...\n", event.ResourceID, *state.LastEvent, *event.LastEvent)

			h := state
			h.TimeTo = *event.EventTime

			if e = d.AddToHistory(h); e != nil {

				l.Warning.Printf("[DB] Something went wrong while saving in the history the previous state. Error: %v\n", e)

			}

			if e = d.Db.Model(&state).Updates(models.State{
				EventTime: event.EventTime,
				LastEvent: event.LastEvent,
				TimeFrom:  *event.EventTime,
			}).Error; e != nil {

				l.Warning.Printf("[DB] Something went wrong while updating the state. Error: %v\n", e)

			} else {

				d.Metrics["count"].With(prometheus.Labels{"type": "States updated"}).Inc()

			}

		}

	}

	eeTotal++
	eeTime += float64(time.Now().UnixNano() - now)

	d.Metrics["time"].With(prometheus.Labels{"type": "Events average registering time"}).Set(eeTime / eeTotal / float64(time.Millisecond))

	d.Metrics["count"].With(prometheus.Labels{"type": "Total events processed"}).Inc()

	d.Metrics["count"].With(prometheus.Labels{"type": "Total " + event.ResourceType + " events processed"}).Inc()

	return

}

// AddToHistory job is to archive the provided state as a new event in the
// history of changes linked to the state provided.
// Parameters:
// - state: State to be archived as a new event in the history of changes of
// the State.
// Returns:
// - e: error raised in case of problems
func (d *DbParameter) AddToHistory(state models.State) (e error) {

	l.Trace.Printf("[DB] Attempting to add the event [ %v ] to the history of changes of the state. Account [ %v ], Resource [ %v ].\n", *state.LastEvent, state.Account, state.ResourceID)

	h := models.Event{
		Account:      state.Account,
		EventTime:    state.EventTime,
		LastEvent:    state.LastEvent,
		MetaData:     state.MetaData,
		Region:       state.Region,
		ResourceID:   state.ResourceID,
		ResourceName: state.ResourceName,
		ResourceType: state.ResourceType,
		TimeFrom:     state.TimeFrom,
		TimeTo:       state.TimeTo,
	}

	if e = d.Db.Create(&h).Error; e != nil {

		l.Warning.Printf("[DB] Something went wrong while adding the event to the history of the state. Error: %v\n", e)

	} else {

		d.Metrics["count"].With(prometheus.Labels{"type": "Histories updated"}).Inc()

	}

	return

}

// DatesWithin job is to check whether the analized time-window overlaps or not,
// reporting how this happens and how much overlapping exists.
// Parameters:
// - EventFrom: int representing the start of the interval for the event.
// - EventTp: int representing the end of the interval for the event.
// - TimeFrom: int representing the start of the time-window.
// - TimeTo: int representing the end of the time-window.
// Returns:
// - inc: const(int) representing the state of the time-window vs the interval.
// - diff: int representing the region of overlapping between both time-windows.
func (d *DbParameter) DatesWithin(EventFrom, EventTo, TimeFrom, TimeTo int64) (inc int, diff int64) {

	l.Trace.Printf("[DB] Attempting to match the time-window with the state/event time interval.\n")

	if EventTo <= TimeFrom {

		l.Debug.Printf("[DB] Time-window is before the time interval of the state/interval.\n")

		inc = notIncluded

		return

	}

	if EventFrom >= TimeTo {

		l.Debug.Printf("[DB] Time-window is after the time interval of the state/interval.\n")

		inc = notIncluded

		return

	}

	if EventFrom <= TimeFrom && TimeTo <= EventTo {

		l.Debug.Printf("[DB] Time-window is completely within the time interval of the state/interval.\n")

		// INNER Time Window
		inc = innerIncluded
		diff = TimeTo - TimeFrom

		return

	}

	if EventFrom <= TimeFrom && EventTo <= TimeTo {

		l.Debug.Printf("[DB] Time-window starts within the time interval of the state/interval and goes forward in the timeline.\n")

		// UPPER Time Window
		inc = upperIncluded
		diff = EventTo - TimeFrom

		return

	}

	if TimeFrom <= EventFrom && TimeTo <= EventTo {

		l.Debug.Printf("[DB] Time-window finished within the time interval of the state/interval and starts before it in the timeline.\n")

		// LOWER Time Window included
		inc = lowerIncluded
		diff = TimeTo - EventFrom

		return

	}

	if TimeFrom <= EventFrom && EventTo <= TimeTo {

		l.Debug.Printf("[DB] The time interval of the state/interval is within the time-windows (AKA starts and finish outside the interval).\n")

		// OUTTER Time Window
		inc = outterIncluded
		diff = EventTo - EventFrom

		return
	}

	l.Warning.Printf("[DB] Time-window is from the 4th dimension, reach the administrator, the CERN, and the whole scientific community ASAP!\n")
	inc = notIncluded

	return

}

// GetAllStates job is to retrieve a snapshot of all the states present in the
// system at the time of invokation-
// Returns:
// - slice of States containing all the resources state in the system
// - error raised in case of problems
func (d *DbParameter) GetAllStates() ([]*models.State, error) {

	l.Trace.Printf("[DB] Attempting to retrieve all the states present in the system.\n")

	var s []*models.State
	var e error

	if e := d.Db.Find(&s).Error; e != nil {

		l.Warning.Printf("[DB] Something went wrong while retrieving the states in the system. Error: %v\n", e)

	}

	return s, e

}

// GetState job is to retrieve the actual state of every resource linked to the
// provided account.
// Parameters:
// - ac: string representing the ID of the account to be processed.
// Returns:
// - slice of State containing the states kept in the system of the resources
// linked to the provided account
// - error raised in case of problems
func (d *DbParameter) GetState(ac string) ([]*models.State, error) {

	l.Trace.Printf("[DB] Attempting to get the actual state of the account [ %v ].\n", ac)

	var s []*models.State
	var e error

	if e := d.Db.Where(&models.State{Account: ac}).Find(&s).Error; e != nil {

		l.Warning.Printf("[DB] Something went wrong while retrieving the state of the account [ %v ]. Error: %v\n", ac, e)

	}

	return s, e

}

// ListState job is to retrieve the list of not terminated states in the system
// for the requested region and resource type..
// Parameters:
// - metric
// - region
// Returns:
// - slice of State containing the states kept in the system.
// - error raised in case of problems
func (d *DbParameter) ListState(metric, region string) ([]*models.MinimalState, error) {

	l.Trace.Printf("[DB] Attempting to get the list of not terminated [ %v ] states for region [ %v ].\n", metric, region)

	var s []*models.MinimalState
	var e error

	filter := "terminated"

	if e := d.Db.Table("states").Where(&models.State{Region: region, ResourceType: metric}).Not(&models.State{LastEvent: &filter}).Order("account").Scan(&s).Error; e != nil {

		l.Warning.Printf("[DB] Something went wrong while retrieving the list of states from the system. Error: %v\n", e)

	}

	return s, e

}

// GetAllHistory job is to retrieve the complete history of changes in the state
// of the provided resource linked to the given account.
// Parameters:
// - ac: string representing the ID of the account to be processed.
// - u: Use reference used as the pattern to match the specific state needed.
// Returns:
// - slice of Event containing the changes in the history of the provided account.
// - error raised in case of problems
func (d *DbParameter) GetAllHistory(ac string, u models.Use) ([]*models.Event, error) {

	l.Trace.Printf("[DB] Attempting to retrieve the full history of the resource [ %v ] from account [ %v ].\n", u.ResourceID, ac)

	var ev []*models.Event
	var e error

	if e := d.Db.Where(&models.Event{
		Account:      ac,
		ResourceID:   u.ResourceID,
		ResourceName: u.ResourceName,
		ResourceType: u.ResourceType,
		MetaData:     u.MetaData,
	}).Find(&ev).Error; e != nil {

		l.Warning.Printf("[DB] Something went wrong while retrieving the history of the resource [ %v ]. Error: %v\n", u.ResourceID, e)

	}

	return ev, e

}

// GetHistory job is to retrieve the history of changes in the state of the
// provided account during the given time-window, with the posibility of filter
// the result by the type of resource.
// Parameters:
// - ac: string representing the ID of the account to be processed.
// - ty: string representing the resource type to use as filter
// - rg: string representing the resource region to use as filter
// - from: int representing the start of the time-window.
// - to: int representing the end of the time-window.
// Returns:
// - slice of Event containing the changes in the history of the provided account.
// - error raised in case of problems
func (d *DbParameter) GetHistory(ac, ty, rg string, from, to int64) ([]*models.Event, error) {

	l.Trace.Printf("[DB] Attempting to retrieve the history of the account [ %v ].\n", ac)

	var ev []*models.Event
	var e error

	whereString := d.Db.NamingStrategy.ColumnName("", "TimeFrom") + " >= ? AND " + d.Db.NamingStrategy.ColumnName("", "TimeTo") + " <= ?"

	if e := d.Db.Where(whereString, from, to).Where(&models.Event{
		Account:      ac,
		ResourceType: ty,
		Region:       rg,
	}).Find(&ev).Error; e != nil {

		l.Warning.Printf("[DB] Something went wrong while retrieving the history of the account [ %v ]. Error: %v\n", ac, e)

	}

	return ev, e

}

// GetUsage job is to compute the usage accumulated by the provided account
// considering the time in the actual state and the history of changes that
// the account resources have had during the provided time-window, with the
// posibility of filtering by resource type.
// Parameters:
// - ac: string representing the ID of the account to be processed.
// - ty: string representing the resource type to use as filter
// - rg: string representing the resource region to use as filter
// - from: int representing the start of the time-window.
// - to: int representing the end of the time-window.
// Returns:
// - Usage reference containing the usage of the account in the system in the
// provided time-window.
// - error raised in case of problems
func (d *DbParameter) GetUsage(ac, ty, rg string, from, to int64) (*models.Usage, error) {

	l.Trace.Printf("[DB] Attempting to compute the usage of the account [ %v ].\n", ac)

	var u models.Usage
	var e error

	// First we check if the actual State is in the time-window
	states, e := d.GetState(ac)

	u.AccountID = ac
	u.TimeFrom = from
	u.TimeTo = to

	for i := range states {

		if ty != "" && states[i].ResourceType != ty {

			l.Trace.Printf("[DB] Resource filtering active. Resource [ %v ] not matching criteria [ %v ].\n", states[i].ResourceType, ty)

			continue

		}

		if rg != "" && states[i].Region != rg {

			l.Trace.Printf("[DB] Resource filtering active. Resource [ %v ] not matching criteria [ %v ].\n", states[i].ResourceType, rg)

			continue

		}

		var use models.Use
		use.Region = states[i].Region
		use.ResourceID = states[i].ResourceID
		use.ResourceName = states[i].ResourceName
		use.ResourceType = states[i].ResourceType
		use.MetaData = states[i].MetaData
		use.Unit = "seconds"

		usage := make(datamodels.JSONdb)

		usage[*states[i].LastEvent] = int64(0)

		// If it is, then we check is the time window goes before the actual state,
		// if not, we create the usage and done
		// If it does..

		if inc, contrib := d.DatesWithin(states[i].TimeFrom, states[i].TimeTo, from, to); inc == innerIncluded || inc == upperIncluded {

			l.Trace.Printf("[DB] Resource [ %v ] state contained within the time-window, skipping history check.\n", states[i].ResourceID)

			usage[*states[i].LastEvent] = usage[*states[i].LastEvent].(int64) + contrib

		} else {

			if inc != notIncluded {

				l.Trace.Printf("[DB] Resource [ %v ] state not (fully) contained within the time-window, checking history of the resource.\n", states[i].ResourceID)

				usage[*states[i].LastEvent] = usage[*states[i].LastEvent].(int64) + contrib

			}

			// We get the History of the State,
			// for each entry we check if the time-windows covers it,
			// if so we add the contribution to the usage
			// if not, we skip it
			history, e := d.GetAllHistory(ac, use)

			if e != nil {

				l.Warning.Printf("[DB] Something went wrong while retrieving the resource [ %v ] history. Error: %v\n", states[i].ResourceID, e)

			} else {

				for id := range history {

					if included, contribution := d.DatesWithin(history[id].TimeFrom, history[id].TimeTo, from, to); included != notIncluded {

						if _, ok := usage[*history[id].LastEvent]; !ok {

							usage[*history[id].LastEvent] = int64(0)

						}

						l.Debug.Printf("[DB] Resource [ %v ] history register contained within the time-window.\n", states[i].ResourceID)

						usage[*history[id].LastEvent] = usage[*history[id].LastEvent].(int64) + contribution

					}

				}

			}

		}

		delete(usage, stateTerminated)

		for id := range usage {

			if usage[id].(int64) == int64(0) {

				delete(usage, id)

			}

		}

		if len(usage) != 0 {

			use.UsageBreakup = usage

			u.Usage = append(u.Usage, &use)

		} else {

			l.Warning.Printf("[DB] The resource [ %v ] has no data in the specified time-window, skipping it's usage report.\n", states[i].ResourceID)

		}

	}

	return &u, e

}

// GetSystemUsage job is to compute the usage accumulated in the whole system
// by account considering the time in the actual state and the history of changes
// that the account resources have had during the provided time-window, with the
// posibility of filtering by resource type.
// Parameters:
// - from: int representing the start of the time-window.
// - to: int representing the end of the time-window.
// - ty: string representing the resource type to use as filter
// - rg: string representing the resource region to use as filter
// Returns:
// - slice of Usages containing the usage per account in the system in the
// provided time-window.
// - error raised in case of problems
func (d *DbParameter) GetSystemUsage(from, to int64, ty, rg string) ([]*models.Usage, error) {

	l.Trace.Printf("[DB] Attempting to compute the system usage.\n")

	var u []*models.Usage
	var e error

	ids := make(map[string]struct{})
	mutex := &sync.Mutex{}

	swg := sizedwaitgroup.New(8)

	if states, e := d.GetAllStates(); e != nil {

		l.Warning.Printf("[DB] Something went wrong while retrieving the states recorded in the system. Error: %v\n", e)

	} else {

		for index := range states {

			// Goroutines start
			swg.Add()
			go func(index int) {

				i := index
				defer swg.Done()

				mutex.Lock()
				_, exists := ids[states[i].Account]
				mutex.Unlock()

				if exists {

					l.Trace.Printf("[DB] Account [ %v ] already processed, skipping...\n", states[i].Account)

					return

				}

				mutex.Lock()
				ids[states[i].Account] = struct{}{}
				mutex.Unlock()

				if su, e := d.GetUsage(states[i].Account, ty, rg, from, to); e != nil || su.Usage == nil {

					l.Warning.Printf("[DB] Something went wrong while retrieving the usage of the account [ %v ]. Error: %v\n", states[i].Account, e)

				} else {

					mutex.Lock()
					u = append(u, su)
					mutex.Unlock()

				}

			}(index)

		}
	}

	swg.Wait()

	return u, e

}

// UpdateStates job is to process the provided states given a pattern, ranging
// through the given states, records the change in the history of the state and
// then updates the state with in the system according to the pattern.
// Parameters:
// - states: slice of State to be updated in the system.
// - pattern: a State reference containing the changes to be done previous to
// the update in the system.
// Returns:
// - error raised in case of problems
func (d *DbParameter) UpdateStates(states []*models.State, pattern models.State) error {

	l.Trace.Printf("[DB] Attempting to update [ %v ] states.\n", len(states))

	var counth, countdb int = 0, 0
	var e error

	for i := range states {

		l.Debug.Printf("[DB] Processing state for resource [ %v ] from account [ %v ].\n", states[i].ResourceID, states[i].Account)

		s := states[i]
		s.TimeTo = pattern.TimeFrom

		if e = d.AddToHistory(*s); e != nil {

			l.Warning.Printf("[DB] Something went wrong while saving the history of the state. Error: %v\n", e)

			counth++

		}

		if e = d.Db.Model(states[i]).Updates(pattern).Error; e != nil {

			l.Warning.Printf("[DB] Something went wrong while updating the state in the system. Error: %v\n", e)

			countdb++

		}

	}

	l.Trace.Printf("[DB] Update of states finished with [ %v ] errors while saving the history and [ %v ] while updating the states\n", counth, countdb)

	return e

}
