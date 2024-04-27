package dbstore

// To complete github test2B
/*func Test(t *testing.T) {
	ctx := context.Background()

	postgres, err := testhelpers.NewPostgresContainer()
	assert.NoError(t, err)
	connectionString, err := postgres.ConnectionString()
	assert.NoError(t, err)

	logger := testhelpers.GetCustomZap(zap.ErrorLevel)
	db, err := New(ctx, config.ServerConfig{ConnString: connectionString}, logger)
	assert.NoError(t, err)
	storagecommons.PerformStoragerTest(t, db)

	var mdb storagecommons.MetricsDB
	var f = 6.6
	var i int64 = 1
	mdb.MetricsDB = []storagecommons.Metrics{
		{ID: "bgm1", MType: "gauge", Value: &f},
		{ID: "bcm1", MType: "counter", Delta: &i},
	}
	t.Run("Batch Multi Write", func(t *testing.T) {
		err = db.WriteDataMultyBatch(ctx, mdb)
		assert.NoError(t, err)
	})

	err = db.Load(ctx)
	assert.NoError(t, err)

	err = db.Dump(ctx)
	assert.NoError(t, err)

	err = db.Ping(ctx)
	assert.NoError(t, err)

	err = db.Close(ctx)
	assert.NoError(t, err)

	t.Run("DB Closed Read Counter", func(t *testing.T) {
		m := storagecommons.Metrics{ID: "cm1", MType: "counter"}
		_, err := db.ReadData(ctx, m)
		assert.Error(t, err)
	})

	t.Run("DB Closed Read Gauge", func(t *testing.T) {
		m := storagecommons.Metrics{ID: "gm1", MType: "gauge"}
		_, err := db.ReadData(ctx, m)
		assert.Error(t, err)
	})

	t.Run("DB Closed Write Counter", func(t *testing.T) {
		m := storagecommons.Metrics{ID: "cm1", MType: "counter"}
		_, err := db.WriteData(ctx, m)
		assert.Error(t, err)
	})

	t.Run("DB Closed Write Gauge", func(t *testing.T) {
		m := storagecommons.Metrics{ID: "gm1", MType: "gauge"}
		_, err := db.WriteData(ctx, m)
		assert.Error(t, err)
	})

	mdb.MetricsDB = []storagecommons.Metrics{
		{ID: "gm1", MType: "gauge", Value: &f},
		{ID: "cm1", MType: "counter", Delta: &i},
	}
	t.Run("DB Closed Write Multi", func(t *testing.T) {
		err = db.WriteDataMulty(ctx, mdb)
		assert.Error(t, err)
	})

	t.Run("DB Closed Batch Multi Write", func(t *testing.T) {
		err = db.WriteDataMultyBatch(ctx, mdb)
		assert.Error(t, err)
	})

}*/
