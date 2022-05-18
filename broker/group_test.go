package broker

import "testing"

func TestGroupAdd(t *testing.T) {

	g := make(Group)

	id := "foo-broker"

	if err := g.AddBroker(Sync("", GetFsSendFunc(id))); err == nil {
		t.Fatal("AddSyncBroker without id should fail")
	}

	if err := g.AddBroker(Sync(id, nil)); err == nil {
		t.Fatal("AddSyncBroker without SendFunc should fail")
	}

	if err := g.AddBroker(Sync(id, GetFsSendFunc(id))); err != nil {
		t.Fatal(err)
	}
}

func TestGroupRm(t *testing.T) {

	g := make(Group)

	id := "foo-broker"
	if err := g.AddBroker(Sync(id, GetFsSendFunc(id))); err != nil {
		t.Fatal(err)
	}

	if err := g.RmBroker(""); err == nil {
		t.Fatal("RmBroker without valid id should fail")
	}

	if err := g.RmBroker(id); err != nil {
		t.Fatal(err)
	}

}
