package mstate_test

import (
	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	. "launchpad.net/gocheck"
	"launchpad.net/juju-core/charm"
	state "launchpad.net/juju-core/mstate"
	"launchpad.net/juju-core/testing"
	"net/url"
	stdtesting "testing"
)

func Test(t *stdtesting.T) { TestingT(t) }

var _ = Suite(&StateSuite{})

type StateSuite struct {
	MgoSuite
	session  *mgo.Session
	machines *mgo.Collection
	charms   *mgo.Collection
	st       *state.State
	ch       charm.Charm
	curl     *charm.URL
}

func (s *StateSuite) SetUpTest(c *C) {
	s.MgoSuite.SetUpTest(c)
	session, err := mgo.Dial(mgoaddr)
	c.Assert(err, IsNil)
	s.session = session

	st, err := state.Dial(mgoaddr)
	c.Assert(err, IsNil)
	s.st = st

	s.machines = session.DB("juju").C("machines")
	s.charms = session.DB("juju").C("charms")

	s.ch = testing.Charms.Dir("dummy")
	url := fmt.Sprintf("local:series/%s-%d", s.ch.Meta().Name, s.ch.Revision())
	s.curl = charm.MustParseURL(url)
}

func (s *StateSuite) TearDownTest(c *C) {
	s.st.Close()
	s.session.Close()
	s.MgoSuite.TearDownTest(c)
}

func (s *StateSuite) TestAddCharm(c *C) {
	// Check that adding charms works correctly.
	bundleURL, err := url.Parse("http://bundle.url")
	c.Assert(err, IsNil)
	dummy, err := s.st.AddCharm(s.ch, s.curl, bundleURL, "dummy-sha256")
	c.Assert(err, IsNil)
	c.Assert(dummy.URL().String(), Equals, s.curl.String())

	mdoc := &struct {
		Url *charm.URL `bson:"_id"`
	}{}
	err = s.charms.Find(bson.D{{"_id", s.curl}}).One(mdoc)
	c.Assert(err, IsNil)
	c.Assert(mdoc.Url, DeepEquals, s.curl)
}

func (s *StateSuite) assertMachineCount(c *C, expect int) {
	ms, err := s.st.AllMachines()
	c.Assert(err, IsNil)
	c.Assert(len(ms), Equals, expect)
}

func (s *StateSuite) TestAllMachines(c *C) {
	numInserts := 42
	for i := 0; i < numInserts; i++ {
		err := s.machines.Insert(bson.D{{"_id", i}})
		c.Assert(err, IsNil)
	}
	s.assertMachineCount(c, numInserts)
	ms, _ := s.st.AllMachines()
	for k, v := range ms {
		c.Assert(v.Id(), Equals, k)
	}
}

func (s *StateSuite) TestAddMachine(c *C) {
	numInserts := 42
	for i := 0; i < numInserts; i++ {
		m, err := s.st.AddMachine()
		c.Assert(err, IsNil)
		c.Assert(m.Id(), Equals, i)
	}
	s.assertMachineCount(c, numInserts)
}

func (s *StateSuite) TestRemoveMachine(c *C) {
	m0, err := s.st.AddMachine()
	c.Assert(err, IsNil)
	m1, err := s.st.AddMachine()
	c.Assert(err, IsNil)
	err = s.st.RemoveMachine(m0.Id())
	c.Assert(err, IsNil)
	s.assertMachineCount(c, 1)
	ms, err := s.st.AllMachines()
	c.Assert(ms[0].Id(), Equals, m1.Id())

	// TODO: Removing a non-existing machine has to fail.
}

func (s *StateSuite) TestMachineInstanceId(c *C) {
	machine, err := s.st.AddMachine()
	c.Assert(err, IsNil)
	err = s.machines.Update(bson.D{{"_id", machine.Id()}}, bson.D{{"instanceid", "spaceship/0"}})
	c.Assert(err, IsNil)

	iid, err := machine.InstanceId()
	c.Assert(err, IsNil)
	c.Assert(iid, Equals, "spaceship/0")
}

func (s *StateSuite) TestMachineSetInstanceId(c *C) {
	machine, err := s.st.AddMachine()
	c.Assert(err, IsNil)
	err = machine.SetInstanceId("umbrella/0")
	c.Assert(err, IsNil)

	n, err := s.machines.Find(bson.D{{"instanceid", "umbrella/0"}}).Count()
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 1)
}

func (s *StateSuite) TestReadMachine(c *C) {
	machine, err := s.st.AddMachine()
	c.Assert(err, IsNil)
	expectedId := machine.Id()
	machine, err = s.st.Machine(expectedId)
	c.Assert(err, IsNil)
	c.Assert(machine.Id(), Equals, expectedId)
}
