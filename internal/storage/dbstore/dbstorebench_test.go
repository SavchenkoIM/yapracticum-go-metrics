package dbstore

import (
	"context"
	"go.uber.org/zap"
	"strconv"
	"testing"
	"yaprakticum-go-track2/internal/config"
	"yaprakticum-go-track2/internal/storage/storagecommons"
	"yaprakticum-go-track2/internal/testhelpers"
)

type dbMultyUpdateFunc func(ctx context.Context, db storagecommons.MetricsDB) error

func performTest(b *testing.B, testNum int) {
	ctx := context.Background()
	db := storagecommons.MetricsDB{}
	for i := 0; i < 500; i++ {
		ii := float64(i)
		elem := storagecommons.Metrics{}
		elem.ID = "gauge" + strconv.Itoa(i)
		elem.MType = "gauge"
		elem.Value = &ii
		db.MetricsDB = append(db.MetricsDB, elem)
	}

	pg, err := testhelpers.NewPostgresContainer()
	if err != nil {
		b.Fatal(err)
	}
	connstr, err := pg.ConnectionString()
	if err != nil {
		b.Fatal(err)
	}
	loger := testhelpers.GetCustomZap(zap.ErrorLevel)
	store, err := New(ctx, config.ServerConfig{ConnString: connstr}, loger)
	if err != nil {
		b.Fatal(err)
	}

	defer func() {
		pg.Close()
	}()

	var fun dbMultyUpdateFunc
	switch testNum {
	case 1:
		fun = store.WriteDataMultyBatch
	case 2:
		fun = store.WriteDataMulty
	default:
		b.Fatal("Unknown test")
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := fun(ctx, db)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDBStore_WriteDataMulty(b *testing.B) {
	performTest(b, 2)
}

func BenchmarkDBStore_WriteDataMultyBatch(b *testing.B) {
	performTest(b, 1)
}
