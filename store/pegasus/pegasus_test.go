package pegasus

import (
	"context"
	"testing"
	"time"

	lib_store "github.com/eko/gocache/lib/v4/store"
	"github.com/smartystreets/assertions/should"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/cast"
)

// run go test -run='TestPegasus*' -race -cover -coverprofile=coverage.txt -covermode=atomic -v ./...
// install local pegasus onebox reference https://pegasus.apache.org/en/docs/build/compile-from-source/
func testPegasusOptions() *OptionsPegasus {
	return &OptionsPegasus{
		MetaServers:       []string{"127.0.0.1:34601", "127.0.0.1:34602", "127.0.0.1:34603"},
		TableName:         "test_pegasus_table",
		TablePartitionNum: 1,
	}
}

func skipPegasusTest(t *testing.T) {
	t.Skip("Pegasus Test skipping because local pegasus onebox not install!")
}

func TestNewPegasus(t *testing.T) {
	Convey("Pegasus TestNewPegasus should return client and nil error", t, func() {
		skipPegasusTest(t)

		ctx := context.Background()

		p, err := NewPegasus(ctx, testPegasusOptions())
		So(err, ShouldBeNil)
		defer p.Close()
	})
}

func Test_validateOptions(t *testing.T) {
	Convey("Pegasus Test validateOptions", t, func() {
		Convey("Test nil validateOptions", func() {
			So(validateOptions(&OptionsPegasus{}), ShouldNotBeNil)
		})
		Convey("Test correct validateOptions", func() {
			So(validateOptions(testPegasusOptions()), ShouldBeNil)
		})
	})
}

func Test_createTable(t *testing.T) {
	Convey("Pegasus Test createTable should return nil", t, func() {
		skipPegasusTest(t)

		ctx := context.Background()

		err := createTable(ctx, testPegasusOptions())
		So(err, ShouldBeNil)
	})
}

func Test_dropTable(t *testing.T) {
	Convey("Pegasus Test dropTable should return nil", t, func() {
		skipPegasusTest(t)

		ctx := context.Background()

		err := dropTable(ctx, testPegasusOptions())
		So(err, ShouldBeNil)
	})
}

func TestPegasusStore_Close(t *testing.T) {
	Convey("Pegasus TestClose for pegasus store", t, func() {
		skipPegasusTest(t)

		ctx := context.Background()

		p, _ := NewPegasus(ctx, testPegasusOptions())
		So(p.Close(), ShouldBeNil)
	})
}

func TestPegasusStore_Get(t *testing.T) {
	Convey("Pegasus TestGet for pegasus store", t, func() {
		skipPegasusTest(t)

		ctx := context.Background()

		p, _ := NewPegasus(ctx, testPegasusOptions())
		defer p.Close()

		k, v := "test-gocache-key", "test-gocache-value"
		p.Set(ctx, k, v)
		value, err := p.Get(ctx, k)
		So(cast.ToString(value), ShouldEqual, v)
		So(err, ShouldBeNil)
	})
}

func TestPegasusStore_GetWithTTL(t *testing.T) {
	Convey("Pegasus TestGetWithTTL for pegasus store", t, func() {
		skipPegasusTest(t)

		ctx := context.Background()

		p, _ := NewPegasus(ctx, testPegasusOptions())
		defer p.Close()

		Convey("test set ttl that not achieve", func() {
			k, v, retention := "test-gocache-key-01", "test-gocache-value", time.Minute*10
			p.Set(ctx, k, v, lib_store.WithExpiration(retention))

			value, ttl, err := p.GetWithTTL(ctx, k)
			So(cast.ToString(value), ShouldEqual, v)
			So(ttl, should.BeLessThanOrEqualTo, retention)
			So(err, ShouldBeNil)
		})
		Convey("test no ttl", func() {
			k, v := "test-gocache-key-02", "test-gocache-value"
			p.Set(ctx, k, v)

			value, ttl, err := p.GetWithTTL(ctx, k)
			So(cast.ToString(value), ShouldEqual, v)
			So(ttl, should.BeLessThanOrEqualTo, PegasusNOTTL)
			So(err, ShouldBeNil)
		})
		Convey("test set ttl that already achieve", func() {
			k, v, retention := "test-gocache-key-03", "test-gocache-value", time.Millisecond*10
			p.Set(ctx, k, v, lib_store.WithExpiration(retention))
			time.Sleep(time.Second * 1)

			value, ttl, err := p.GetWithTTL(ctx, k)
			So(cast.ToString(value), ShouldBeEmpty)
			So(ttl, should.BeLessThanOrEqualTo, PegasusNOENTRY)
			So(err, ShouldBeNil)
		})
	})
}

func TestPegasusStore_Set(t *testing.T) {
	Convey("Pegasus TestSet for pegasus store", t, func() {
		skipPegasusTest(t)

		ctx := context.Background()

		p, _ := NewPegasus(ctx, testPegasusOptions())
		defer p.Close()

		k, v := "test-gocache-key", "test-gocache-value"
		err := p.Set(ctx, k, v)
		So(err, ShouldBeNil)
	})
}

func TestPegasusStore_setTags(t *testing.T) {
	Convey("Pegasus Test set tags for pegasus store", t, func() {
		skipPegasusTest(t)

		ctx := context.Background()

		p, _ := NewPegasus(ctx, testPegasusOptions())
		defer p.Close()

		k, tags := "test-gocache-tags-key", []string{"test01", "test02"}
		err := p.setTags(ctx, k, tags)
		So(err, ShouldBeNil)
	})
}

func TestPegasusStore_Delete(t *testing.T) {
	Convey("Pegasus TestDelete for pegasus store", t, func() {
		skipPegasusTest(t)

		ctx := context.Background()

		p, _ := NewPegasus(ctx, testPegasusOptions())
		defer p.Close()

		k, v := "test-gocache-key", "test-gocache-value"
		p.Set(ctx, k, v)

		err := p.Delete(ctx, k)
		So(err, ShouldBeNil)
	})
}

func TestPegasusStore_Invalidate(t *testing.T) {
	Convey("Pegasus TestInvalidate for pegasus store", t, func() {
		skipPegasusTest(t)

		ctx := context.Background()

		p, _ := NewPegasus(ctx, testPegasusOptions())
		defer p.Close()

		err := p.Invalidate(ctx)
		So(err, ShouldBeNil)
	})
}

func TestPegasusStore_Clear(t *testing.T) {
	Convey("Pegasus TestClear for pegasus store", t, func() {
		skipPegasusTest(t)

		ctx := context.Background()

		p, _ := NewPegasus(ctx, testPegasusOptions())
		defer p.Close()

		k1, v1 := "test-gocache-key-01", "test-gocache-value"
		k2, v2 := "test-gocache-key-01", "test-gocache-value"
		p.Set(ctx, k1, v1)
		p.Set(ctx, k2, v2)

		err := p.Clear(ctx)
		So(err, ShouldBeNil)
	})
}
