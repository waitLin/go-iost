package blockcache

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/core/tx"
	"github.com/iost-official/prototype/db"
	"github.com/iost-official/prototype/vm"
	"github.com/iost-official/prototype/vm/lua"
	. "github.com/smartystreets/goconvey/convey"
)

func TestStdTxsVerifier(t *testing.T) {
	dbx, err := db.DatabaseFactory("redis")
	if err != nil {
		panic(err.Error())
	}
	sdb := state.NewDatabase(dbx)
	pool := state.NewPool(sdb)
	main := lua.NewMethod(2, "main", 0, 1)
	code := `function main()
				Put("hello", "world")
				return "success"
			end`

	txs := make([]*tx.Tx, 0)

	pool.PutHM("iost", "a", state.MakeVFloat(100000))

	for j := 0; j < 100; j++ {
		lc := lua.NewContract(vm.ContractInfo{Prefix: strconv.Itoa(j), GasLimit: 10000, Price: 1, Publisher: vm.IOSTAccount("a")}, code, main)
		txx := tx.NewTx(int64(j), &lc)
		txs = append(txs, &txx)
	}

	fmt.Println("len", len(txs))
	//fmt.Println(txs[0].contract)
	p2, i, err := StdTxsVerifier(txs, pool)
	fmt.Println(i)
	fmt.Println("error", err.Error())
	fmt.Println(p2.GetHM("iost", "a"))
	//fmt.Println(p2.GetPatch())
	//fmt.Println(p2.Get("mark"))
	p := p2.(*state.PoolImpl)
	count := 0
	for p != nil {
		p = p.Parent()
		count++
	}
	fmt.Println("depth:", count)
}

func TestStdBlockVerifier(t *testing.T) {

}

func TestStdCacheVerifier(t *testing.T) {
	Convey("Test of StdCacheVerifier", t, func() {
		dbx, err := db.DatabaseFactory("redis")
		So(err, ShouldBeNil)
		sdb := state.NewDatabase(dbx)
		pool := state.NewPool(sdb)
		main := lua.NewMethod(2, "main", 0, 1)
		code := `function main()
				Put("hello", "world")
				return "success"
			end`

		pool.PutHM("iost", "a", state.MakeVFloat(100000))

		for j := 0; j < 90; j++ {
			lc := lua.NewContract(vm.ContractInfo{Prefix: strconv.Itoa(j), GasLimit: 10000, Price: 1, Publisher: vm.IOSTAccount("a")}, code, main)
			txx := tx.NewTx(int64(j), &lc)
			StdCacheVerifier(&txx, pool, &vm.Context{})
		}
		fmt.Println(pool.GetHM("iost", "a"))
		p := pool.(*state.PoolImpl)
		count := 0
		for p != nil {
			p = p.Parent()
			count++
		}
		So(count, ShouldEqual, 1)
	})

	Convey("Test of context", t, func() {
		dbx, err := db.DatabaseFactory("redis")
		So(err, ShouldBeNil)
		sdb := state.NewDatabase(dbx)
		pool := state.NewPool(sdb)
		main := lua.NewMethod(2, "main", 0, 1)
		code := `function main()
				if (Random(0.5))
				then 
					Transfer("a", "b", 10000)
				end
				return "success"
			end`

		pool.PutHM("iost", "a", state.MakeVFloat(100000))

		lc := lua.NewContract(vm.ContractInfo{Prefix: "ahaha", GasLimit: 10000, Price: 1, Publisher: vm.IOSTAccount("a")}, code, main)
		txx := tx.NewTx(1, &lc)
		err = StdCacheVerifier(&txx, pool, &vm.Context{ParentHash: []byte{1}})
		So(err, ShouldBeNil)

		balance, err := pool.GetHM("iost", "a")
		So(err, ShouldBeNil)

		So(balance.(*state.VFloat).ToFloat64() > 90000, ShouldBeTrue)

		err = StdCacheVerifier(&txx, pool, &vm.Context{})
		So(err, ShouldBeNil)
		balance, err = pool.GetHM("iost", "a")
		So(err, ShouldBeNil)
		So(balance.(*state.VFloat).ToFloat64() < 90000, ShouldBeTrue)
	})
}

func BenchmarkStdTxsVerifier(b *testing.B) {
	dbx, err := db.DatabaseFactory("redis")
	if err != nil {
		panic(err.Error())
	}
	sdb := state.NewDatabase(dbx)
	pool := state.NewPool(sdb)
	main := lua.NewMethod(2, "main", 0, 1)
	code := `function main()
				Put("hello", "world")
				return "success"
			end`

	txs := make([]*tx.Tx, 0)
	for j := 0; j < 10000; j++ {
		lc := lua.NewContract(vm.ContractInfo{Prefix: strconv.Itoa(j), GasLimit: 10000, Price: 0, Publisher: vm.IOSTAccount("a")}, code, main)
		txx := tx.NewTx(int64(j), &lc)
		txs = append(txs, &txx)
	}

	//fmt.Println(txs[500].contract)
	//var k int
	for i := 0; i < b.N; i++ {
		pool, _, _ = StdTxsVerifier(txs, pool)
	}
	//fmt.Println(pool.GetPatch())

	fmt.Println(pool.GetHM("iost", "a"))

}

func BenchmarkStdCacheVerifier(b *testing.B) {
	dbx, err := db.DatabaseFactory("redis")
	if err != nil {
		panic(err.Error())
	}
	sdb := state.NewDatabase(dbx)
	pool := state.NewPool(sdb)
	main := lua.NewMethod(2, "main", 0, 1)
	code := `function main()
				Put("hello", "world")
				return "success"
			end`

	pool.PutHM("iost", "a", state.MakeVFloat(1000000000))

	lc := lua.NewContract(vm.ContractInfo{Prefix: "ahaha", GasLimit: 10000, Price: 1, Publisher: vm.IOSTAccount("a")}, code, main)
	txx := tx.NewTx(123, &lc)
	var p2 state.Pool
	for i := 0; i < b.N; i++ {
		p2 = pool
		b.StartTimer()
		for j := 0; j < 10000; j++ {
			//err := StdCacheVerifier(&txx, p2)
			p2, _, err = StdTxsVerifier([]*tx.Tx{&txx}, p2)
			if err != nil {
				fmt.Println(err)
			}
		}
		b.StopTimer()
		p := p2.(*state.PoolImpl)
		count := 0
		for p != nil {
			p = p.Parent()
			count++
		}
		fmt.Println("depth:", count)
		fmt.Println(p2.GetHM("iost", "a"))
	}

}