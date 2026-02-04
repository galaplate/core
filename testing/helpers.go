package testing

type WithRefreshDatabase struct {
	TestCase
}

func (w *WithRefreshDatabase) SetupTest() {
	w.EnableRefreshDatabase()
	w.TestCase.SetupTest()
}

func (w *WithRefreshDatabase) SetupSuite() {
	w.EnableRefreshDatabase()
}

type RefreshDatabaseBeforeEachTest struct {
	TestCase
}

func (r *RefreshDatabaseBeforeEachTest) SetupTest() {
	r.EnableRefreshDatabase()
	r.RefreshDatabaseBetweenTests()
	r.TestCase.SetupTest()
}
