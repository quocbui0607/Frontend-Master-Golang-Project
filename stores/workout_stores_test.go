package stores

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T, ctx context.Context) (*sql.DB, error) {
	env, _ := ctx.Value("env").(map[string]string)

	db, err := sql.Open("pgx", fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		env["DB_HOST"], env["DB_USERNAME"], env["DB_PASSWORD"], env["DB_NAME"], env["DB_PORT"], env["SSL_MODE"]))
	if err != nil {
		t.Fatalf("opening test db: %v", err)
	}

	// run the migratoins for our test db
	err = Migrate(db, "../migrations/")
	if err != nil {
		t.Fatalf("migrating test db error: %v", err)
	}

	_, err = db.Exec(`TRUNCATE users, workouts, workout_entries CASCADE`)
	if err != nil {
		t.Fatalf("truncating tables %v", err)
	}

	return db, nil
}

func TestCreateWorkout(t *testing.T) {
	envData, err := godotenv.Read("../.env")
	if err != nil {
		t.Fatalf("Không thể đọc file .env: %v", err)
	}

	ctx := context.WithValue(context.Background(), "env", envData)
	db, _ := setupTestDB(t, ctx)
	defer db.Close()

	store := NewPostgresWorkoutStore(db)
	userStore := NewPostgresUserStore(db)

	testUser := &User{
		Username: "test",
		Email:    "test@example.com",
	}

	err = testUser.PasswordHash.Set("securepassword")
	require.NoError(t, err)

	err = userStore.CreateUser(testUser)
	require.NoError(t, err)

	tests := []struct {
		name    string
		workout *Workout
		wantErr bool
	}{
		{
			name: "valid workout",
			workout: &Workout{
				UserID:          testUser.ID,
				Title:           "push day",
				Description:     "upper body day",
				DurationMinutes: 60,
				CaloriesBurned:  200,
				Entries: []WorkoutEntry{
					{
						ExerciseName: "Bench press",
						Sets:         3,
						Reps:         new(10),
						Weight:       new(135.5),
						Notes:        "warm up properly",
						OrderIndex:   1,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "workout with invalid entries",
			workout: &Workout{
				UserID:          testUser.ID,
				Title:           "full body",
				Description:     "complete workout",
				DurationMinutes: 90,
				CaloriesBurned:  500,
				Entries: []WorkoutEntry{
					{
						ExerciseName: "Plank",
						Sets:         3,
						Reps:         new(60),
						Notes:        "keep form",
						OrderIndex:   1,
					},
					{
						ExerciseName:    "squats",
						Sets:            4,
						Reps:            new(12),
						DurationSeconds: new(60),
						Weight:          new(185.0),
						Notes:           "full depth",
						OrderIndex:      2,
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createdWorkout, err := store.CreateWorkout(tt.workout)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.workout.Title, createdWorkout.Title)
			assert.Equal(t, tt.workout.Description, createdWorkout.Description)
			assert.Equal(t, tt.workout.DurationMinutes, createdWorkout.DurationMinutes)

			retrieved, err := store.GetWorkoutByID(int64(createdWorkout.ID))
			require.NoError(t, err)

			assert.Equal(t, createdWorkout.ID, retrieved.ID)
			assert.Equal(t, len(tt.workout.Entries), len(retrieved.Entries))

			for i := range retrieved.Entries {
				assert.Equal(t, tt.workout.Entries[i].ExerciseName, retrieved.Entries[i].ExerciseName)
				assert.Equal(t, tt.workout.Entries[i].Sets, retrieved.Entries[i].Sets)
				assert.Equal(t, tt.workout.Entries[i].OrderIndex, retrieved.Entries[i].OrderIndex)
			}

		})
	}
}

func IntPtr(i int) *int {
	return &i
}

func FloatPtr(i float64) *float64 {
	return &i
}
