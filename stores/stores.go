package stores

import (
	"database/sql"
	"fmt"
	"quocbui0607/femProject/utils"
)

type Workout struct {
	ID              int            `json:"id"`
	UserID          int            `json:"user_id"`
	Title           string         `json:"title"`
	Description     string         `json:"description"`
	DurationMinutes int            `json:"duration_minutes"`
	CaloriesBurned  int            `json:"calories_burned"`
	Entries         []WorkoutEntry `json:"entries"`
}

type WorkoutEntry struct {
	ID              int      `json:"id"`
	ExerciseName    string   `json:"exercise_name"`
	Sets            int      `json:"sets"`
	Reps            *int     `json:"reps"`
	DurationSeconds *int     `json:"duration_seconds"`
	Weight          *float64 `json:"weight"`
	Notes           string   `json:"notes"`
	OrderIndex      int      `json:"order_index"`
}

type PostgresWorkoutStore struct {
	db *sql.DB
}

func NewPostgresWorkoutStore(db *sql.DB) *PostgresWorkoutStore {
	return &PostgresWorkoutStore{db: db}
}

type WorkoutStore interface {
	CreateWorkout(*Workout) (*Workout, error)
	GetWorkout(id int64) (*Workout, error)
}

func (pg *PostgresWorkoutStore) CreateWorkout(wk *Workout) (*Workout, error) {
	tx, err := pg.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("%s_Error Begin: %v", utils.WORKOUT_PREFIX, err)
	}

	defer tx.Rollback()

	query := `INSERT INTO workout (title,description,duration_minutes, calories_burned)
	VALUES ($1, $2, $3, $4)
	RETURNING id
	`

	err = tx.QueryRow(query, wk.Title, wk.Description, wk.DurationMinutes, wk.CaloriesBurned).Scan(&wk.ID)
	if err != nil {
		return nil, fmt.Errorf("%s_Error QueryRow: %v", utils.WORKOUT_PREFIX, err)
	}

	for _, entry := range wk.Entries {
		query := `
		INSERT INTO workout_entries (workout_id,exercise_name,sets, reps, duration_seconds, weight, notes, order_index)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
		`
		err = tx.QueryRow(query,
			wk.ID,
			entry.ID,
			entry.ExerciseName,
			entry.Sets, entry.Reps, entry.DurationSeconds, entry.Weight, entry.Notes, entry.OrderIndex).Scan(&entry.ID)
		if err != nil {
			return nil, fmt.Errorf("%s_Error QueryRow entry: %v", utils.WORKOUT_PREFIX, err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("%s_Error Commit: %v", utils.WORKOUT_PREFIX, err)
	}

	return wk, nil
}

func (pg *PostgresWorkoutStore) GetWorkout(id int64) (*Workout, error) {
	wk := &Workout{}
	return wk, nil
}
