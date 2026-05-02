package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/domain"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func loadJSON[T any](path string) ([]T, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out []T
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func main() {
	_ = godotenv.Load()

	dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dsn == "" {
		dsn = strings.TrimSpace(os.Getenv("CONNECTION_STRING"))
	}
	if dsn == "" {
		log.Fatal("DATABASE_URL or CONNECTION_STRING not set. Example: export DATABASE_URL=postgres://user:pass@host:5432/dbname?sslmode=disable")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed connect db: %v", err)
	}

	if err := db.AutoMigrate(
		&domain.Category{},
		&domain.User{},
		&domain.EmployerProfile{},
		&domain.WorkerProfile{},
		&domain.WorkerCV{},
		&domain.Availability{},
		&domain.Job{},
		&domain.JobSchedule{},
		&domain.Application{},
	); err != nil {
		log.Fatalf("migrate failed: %v", err)
	}

	tx := db.Begin()
	if tx.Error != nil {
		log.Fatalf("begin tx: %v", tx.Error)
	}

	insert := func(v any) error {
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(v).Error; err != nil {
			return err
		}
		return nil
	}

	if cats, err := loadJSON[domain.Category]("internal/seeder/categories.json"); err == nil && len(cats) > 0 {
		if err := insert(&cats); err != nil {
			log.Fatalf("insert categories: %v", err)
		}
		fmt.Printf("seeded %d categories\n", len(cats))
	}

	if us, err := loadJSON[domain.User]("internal/seeder/users.json"); err == nil && len(us) > 0 {
		if err := insert(&us); err != nil {
			log.Fatalf("insert users: %v", err)
		}
		fmt.Printf("seeded %d users\n", len(us))
	}

	if emps, err := loadJSON[domain.EmployerProfile]("internal/seeder/employers.json"); err == nil && len(emps) > 0 {
		if err := insert(&emps); err != nil {
			log.Fatalf("insert employers: %v", err)
		}
		fmt.Printf("seeded %d employers\n", len(emps))
	}

	if ws, err := loadJSON[domain.WorkerProfile]("internal/seeder/workers.json"); err == nil && len(ws) > 0 {
		for i := range ws {
			if ws[i].ID == uuid.Nil {
				ws[i].ID = uuid.New()
			}
		}
		if err := insert(&ws); err != nil {
			log.Fatalf("insert workers: %v", err)
		}
		fmt.Printf("seeded %d workers\n", len(ws))
	}

	if cvs, err := loadJSON[domain.WorkerCV]("internal/seeder/cvs.json"); err == nil && len(cvs) > 0 {
		if err := insert(&cvs); err != nil {
			log.Fatalf("insert cvs: %v", err)
		}
		fmt.Printf("seeded %d cvs\n", len(cvs))
	}

	if av, err := loadJSON[domain.Availability]("internal/seeder/availabilities.json"); err == nil && len(av) > 0 {
		if err := insert(&av); err != nil {
			log.Fatalf("insert availabilities: %v", err)
		}
		fmt.Printf("seeded %d availabilities\n", len(av))
	}

	if jobs, err := loadJSON[domain.Job]("internal/seeder/jobs.json"); err == nil && len(jobs) > 0 {
		if err := insert(&jobs); err != nil {
			log.Fatalf("insert jobs: %v", err)
		}
		fmt.Printf("seeded %d jobs\n", len(jobs))
	}

	if js, err := loadJSON[domain.JobSchedule]("internal/seeder/job_schedules.json"); err == nil && len(js) > 0 {
		if err := insert(&js); err != nil {
			log.Fatalf("insert job_schedules: %v", err)
		}
		if err := tx.Exec("SELECT setval(pg_get_serial_sequence('job_schedules','id'), COALESCE((SELECT MAX(id) FROM job_schedules), 0))").Error; err != nil {
			log.Fatalf("reset job_schedules id sequence: %v", err)
		}
		fmt.Printf("seeded %d job_schedules\n", len(js))
	}

	if apps, err := loadJSON[domain.Application]("internal/seeder/applications.json"); err == nil && len(apps) > 0 {
		if err := insert(&apps); err != nil {
			log.Fatalf("insert applications: %v", err)
		}
		fmt.Printf("seeded %d applications\n", len(apps))
	}

	if err := tx.Commit().Error; err != nil {
		log.Fatalf("commit tx: %v", err)
	}

	fmt.Println("seeding complete")
}
