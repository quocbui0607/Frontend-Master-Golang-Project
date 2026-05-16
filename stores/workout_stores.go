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
	GetWorkoutByID(id int64) (*Workout, error)
	UpdateWorkoutByID(*Workout) error
	DeleteWorkoutByID(id int64) error
	GetWorkoutOwner(id int64) (int, error)
}

func (pg *PostgresWorkoutStore) CreateWorkout(wk *Workout) (*Workout, error) {
	tx, err := pg.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("%s_Error Begin: %v", utils.WORKOUT_PREFIX, err)
	}

	defer tx.Rollback()

	query := `INSERT INTO workouts (user_id, title, description, duration_minutes, calories_burned)
  VALUES ($1, $2, $3, $4, $5)
  RETURNING id 
	`

	err = tx.QueryRow(query, wk.UserID, wk.Title, wk.Description, wk.DurationMinutes, wk.CaloriesBurned).Scan(&wk.ID)
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

func (pg *PostgresWorkoutStore) GetWorkoutByID(id int64) (*Workout, error) {
	wk := &Workout{}
	query := `
	SELECT id, title, description, duration_minutes, calories_burned
	FROM workouts
	WHERE id = $1`

	err := pg.db.QueryRow(query, id).Scan(&wk.ID, &wk.Title,
		&wk.Description, &wk.DurationMinutes, &wk.CaloriesBurned)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	entryQuery := `
	SELECT id, exercise_name, sets, reps, duration_seconds, weight, notes, order_index
	FROM workout_entries
	WHERE workout_id = $1
	ORDER BY order_index
	`
	rows, err := pg.db.Query(entryQuery, id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var entry WorkoutEntry
		err = rows.Scan(
			&entry.ID,
			&entry.ExerciseName,
			&entry.Sets,
			&entry.Reps,
			&entry.DurationSeconds,
			&entry.Weight,
			&entry.Notes,
			&entry.OrderIndex,
		)
		if err != nil {
			return nil, err
		}

		wk.Entries = append(wk.Entries, entry)

	}

	return wk, nil
}

func (pg *PostgresWorkoutStore) UpdateWorkoutByID(workout *Workout) error {
	tx, err := pg.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
	UPDATE workouts
	SET title = $1, description = $2, duration_minutes = $3, calories_burned = $4
	WHERE id = $5`

	result, err := tx.Exec(query, workout.Title,
		workout.Description,
		workout.DurationMinutes,
		workout.CaloriesBurned, workout.ID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	_, err = tx.Exec(`DELETE FROM workout_entries WHERE workout_id=$1`, workout.ID)
	if err != nil {
		return err
	}

	for _, entry := range workout.Entries {
		query := `
		INSERT INTO workout_entries (workout_id, exercise_name, sets, reps, duration_seconds, weight, notes, order_index)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

		_, err := tx.Exec(query, workout.ID,
			entry.ExerciseName,
			entry.Sets,
			entry.Reps,
			entry.DurationSeconds,
			entry.Weight,
			entry.Notes,
			entry.OrderIndex)
		if err != nil {
			return err
		}

	}

	return tx.Commit()
}

func (pg *PostgresWorkoutStore) DeleteWorkoutByID(id int64) error {
	query := `
		DELETE from workouts
		WHERE id = $1`

	result, err := pg.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (pg *PostgresWorkoutStore) GetWorkoutOwner(workoutID int64) (int, error) {
	var userID int

	query := `
		SELECT user_id
		FROM workouts
		WHERE id = $1
  	`

	err := pg.db.QueryRow(query, workoutID).Scan(&userID)
	if err != nil {
		return 0, err
	}

	return userID, nil
}
