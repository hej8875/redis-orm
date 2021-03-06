package model_test

import (
	"fmt"
	"time"

	. "github.com/ezbuy/redis-orm/example/model"
	"github.com/ezbuy/redis-orm/orm"

	"log"

	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("manager", func() {
	MySQLSetup(&MySQLConfig{
		Host:     "localhost",
		Port:     3306,
		UserName: "ezorm_user",
		Password: "ezorm_pass",
		Database: "ezorm",
	})

	RedisSetUp(&RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Password: "",
	})

	BeforeEach(func() {
		tx, err := MySQL().BeginTx()
		Ω(err).ShouldNot(HaveOccurred())
		defer tx.Close()

		users := []*User{}
		for i := 0; i < 100; i++ {
			user := UserMgr.NewUser()
			user.Name = fmt.Sprintf("name%d", i)
			user.Mailbox = fmt.Sprintf("name%d@ezbuy.com", i)
			user.HeadUrl = fmt.Sprintf("name%d.png", i)
			user.Password = fmt.Sprintf("pwd%d", i)
			if i%2 == 0 {
				user.Sex = true
			} else {
				user.Sex = false
			}
			user.Age = int32(i)
			user.CreatedAt = time.Now()
			user.UpdatedAt = user.CreatedAt
			user.Longitude = 103.754
			user.Latitude = 1.3282
			users = append(users, user)
		}
		n, err := UserDBMgr(tx).BatchCreate(users)
		Ω(n).To(Equal(int64(100)))
		Ω(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		tx, err := MySQL().BeginTx()
		Ω(err).ShouldNot(HaveOccurred())
		defer tx.Close()
		n, err := UserDBMgr(tx).DeleteBySQL("")
		Ω(n).To(Equal(int64(100)))
		Ω(err).ShouldNot(HaveOccurred())
	})

	Describe("load", func() {
		It("mysql => redis", func() {
			Ω(MySQL()).ShouldNot(BeNil())
			dbMgr := UserDBMgr(MySQL())
			Ω(UserRedisMgr(Redis()).Load(dbMgr)).ShouldNot(HaveOccurred())
		})
	})

	It("vset", func() {
		vset := orm.NewVSet()
		Ω(vset).ShouldNot(BeNil())

		vset.Add(101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115, 116)
		Ω(vset.Values(1, 0, 6)).Should(Equal([]interface{}{101, 102, 103, 104, 105, 106}))
		Ω(vset.Values(1, 6, 6)).Should(Equal([]interface{}{107, 108, 109, 110, 111, 112}))
		Ω(vset.Values(1, 12, -1)).Should(Equal([]interface{}{113, 114, 115, 116}))
		vset.SortAdd(2, 103, 102, 105, 108, 113, 112)
		Ω(vset.Values(2, 0, 2)).Should(Equal([]interface{}{103, 102}))
		Ω(vset.Values(2, 2, 2)).Should(Equal([]interface{}{105, 108}))
		Ω(vset.Values(2, 4, -1)).Should(Equal([]interface{}{113, 112}))
	})
})

var _ = Describe("redis-orm.mysql", func() {
	Describe("CRUD", func() {
		It("create", func() {
			MySQLSetup(&MySQLConfig{
				Host:     "localhost",
				Port:     3306,
				UserName: "ezorm_user",
				Password: "ezorm_pass",
				Database: "ezorm",
			})

			user := UserMgr.NewUser()
			user.Name = "user01"
			user.Mailbox = "user01@sss.fff"
			user.HeadUrl = "aaaa.png"
			user.Password = "123456"
			user.CreatedAt = time.Now()
			user.UpdatedAt = user.CreatedAt
			user.Longitude = 103.754
			user.Latitude = 1.3282

			tx, err := MySQL().BeginTx()
			Ω(err).ShouldNot(HaveOccurred())
			defer tx.Close()

			mgr := UserDBMgr(tx)
			user.Name = ""
			log.Println("user validator =>", user.Validate())

			//! create
			n, err := mgr.Create(user)
			Ω(n).To(Equal(int64(1)))
			log.Println("mysql.tx.crud.create =>", n, err)
			Ω(err).ShouldNot(HaveOccurred())

			log.Println("create User :", user)

			//! update
			user.HeadUrl = "bbbb.png"
			user.UpdatedAt = time.Now()
			n, err = mgr.Update(user)
			Ω(n).To(Equal(int64(1)))
			Ω(err).ShouldNot(HaveOccurred())

			//! fetch check
			obj, err := mgr.Fetch(user.GetPrimaryKey())
			Ω(err).ShouldNot(HaveOccurred())
			Ω(obj.HeadUrl).To(Equal(user.HeadUrl))

			exist, err := mgr.Exist(user.GetPrimaryKey())
			Ω(err).ShouldNot(HaveOccurred())
			Ω(exist).To(Equal(true))
			//! delete
			n, err = mgr.Delete(user)
			log.Println("mysql.tx.crud.delete =>", n, err)
			Ω(n).To(Equal(int64(1)))
			Ω(err).ShouldNot(HaveOccurred())

			//! fetch check
			_, err = mgr.Fetch(user.GetPrimaryKey())
			Ω(err).Should(HaveOccurred())

			//! save

			user.HeadUrl = "ccc.png"
			n, err = mgr.Save(user)
			Ω(n).To(Equal(int64(1)))
			Ω(err).ShouldNot(HaveOccurred())

			n, err = mgr.Delete(user)
			Ω(n).To(Equal(int64(1)))
			Ω(err).ShouldNot(HaveOccurred())

		})

	})
})

var _ = Describe("redis-orm.mysql", func() {
	BeforeEach(func() {
		MySQLSetup(&MySQLConfig{
			Host:     "localhost",
			Port:     3306,
			UserName: "ezorm_user",
			Password: "ezorm_pass",
			Database: "ezorm",
		})
		d, _ := time.ParseDuration("6ms")
		MySQL().SlowLog(d)
	})

	Describe("CRUD", func() {
		It("user crud test", func() {
			user := UserMgr.NewUser()
			user.Name = "user01"
			user.Mailbox = "user01@sss.fff"
			user.HeadUrl = "aaaa.png"
			user.Password = "123456"
			user.CreatedAt = time.Now()
			user.UpdatedAt = user.CreatedAt
			user.Longitude = 103.754
			user.Latitude = 1.3282

			tx, err := MySQL().BeginTx()
			Ω(err).ShouldNot(HaveOccurred())
			defer tx.Close()

			mgr := UserDBMgr(tx)
			user.Name = ""
			log.Println("user validator =>", user.Validate())

			//! create
			n, err := mgr.Create(user)
			Ω(n).To(Equal(int64(1)))
			log.Println("mysql.tx.crud.create =>", n, err)
			Ω(err).ShouldNot(HaveOccurred())

			log.Println("create User :", user)

			//! update
			user.HeadUrl = "bbbb.png"
			user.UpdatedAt = time.Now()
			n, err = mgr.Update(user)
			Ω(n).To(Equal(int64(1)))
			Ω(err).ShouldNot(HaveOccurred())

			//! fetch check
			obj, err := mgr.Fetch(user.GetPrimaryKey())
			Ω(err).ShouldNot(HaveOccurred())
			Ω(obj.HeadUrl).To(Equal(user.HeadUrl))

			//! delete
			n, err = mgr.Delete(user)
			log.Println("mysql.tx.crud.delete =>", n, err)
			Ω(n).To(Equal(int64(1)))
			Ω(err).ShouldNot(HaveOccurred())

			//! fetch check
			_, err = mgr.Fetch(user.GetPrimaryKey())
			Ω(err).Should(HaveOccurred())

			//! save

			user.HeadUrl = "ccc.png"
			n, err = mgr.Save(user)
			Ω(n).To(Equal(int64(1)))
			Ω(err).ShouldNot(HaveOccurred())

			n, err = mgr.Delete(user)
			Ω(n).To(Equal(int64(1)))
			Ω(err).ShouldNot(HaveOccurred())
		})

		Measure("mysql.bench", func(b Benchmarker) {
			b.Time("crud.runtime", func() {
				user := UserMgr.NewUser()
				user.Name = "user01"
				user.Mailbox = "user01@sss.fff"
				user.HeadUrl = "aaaa.png"
				user.Password = "123456"
				user.CreatedAt = time.Now()
				user.UpdatedAt = user.CreatedAt
				user.Longitude = 103.754
				user.Latitude = 1.3282
				tx, err := MySQL().BeginTx()
				Ω(err).ShouldNot(HaveOccurred())
				defer tx.Close()

				mgr := UserDBMgr(tx)

				//! create
				n, err := mgr.Create(user)
				Ω(n).To(Equal(int64(1)))
				Ω(err).ShouldNot(HaveOccurred())

				//! update
				user.HeadUrl = "bbbb.png"
				user.UpdatedAt = time.Now()
				n, err = mgr.Update(user)
				Ω(n).To(Equal(int64(1)))
				Ω(err).ShouldNot(HaveOccurred())

				//! fetch check
				obj, err := mgr.Fetch(user.GetPrimaryKey())
				Ω(err).ShouldNot(HaveOccurred())
				Ω(obj.HeadUrl).To(Equal(user.HeadUrl))

				//! delete
				n, err = mgr.Delete(user)
				Ω(n).To(Equal(int64(1)))
				Ω(err).ShouldNot(HaveOccurred())

				//! fetch check
				_, err = mgr.Fetch(user.GetPrimaryKey())
				Ω(err).Should(HaveOccurred())

				//! save
				user.HeadUrl = "ccc.png"
				n, err = mgr.Save(user)
				Ω(n).To(Equal(int64(1)))
				Ω(err).ShouldNot(HaveOccurred())

				n, err = mgr.Delete(user)
				Ω(n).To(Equal(int64(1)))
				Ω(err).ShouldNot(HaveOccurred())
			})
		}, 1)
	})

	Describe("Finder", func() {

		BeforeEach(func() {

			tx, err := MySQL().BeginTx()
			Ω(err).ShouldNot(HaveOccurred())
			defer tx.Close()

			mgr := UserDBMgr(tx)

			users := []*User{}
			for i := 0; i < 100; i++ {
				user := UserMgr.NewUser()
				user.Name = fmt.Sprintf("name%d", i)
				user.Mailbox = fmt.Sprintf("name%d@ezbuy.com", i)
				user.HeadUrl = fmt.Sprintf("name%d.png", i)
				user.Password = fmt.Sprintf("pwd%d", i)
				if i%2 == 0 {
					user.Sex = true
				} else {
					user.Sex = false
				}
				user.Age = int32(i)
				user.CreatedAt = time.Now()
				user.UpdatedAt = user.CreatedAt
				user.Longitude = 103.754
				user.Latitude = 1.3282
				users = append(users, user)
			}

			n, err := mgr.BatchCreate(users)
			Ω(n).To(Equal(int64(100)))
			Ω(err).ShouldNot(HaveOccurred())

		})
		AfterEach(func() {
			tx, err := MySQL().BeginTx()
			Ω(err).ShouldNot(HaveOccurred())
			defer tx.Close()

			mgr := UserDBMgr(tx)
			_, err = mgr.DeleteBySQL("")
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("unique", func() {
			unique := &MailboxPasswordOfUserUK{
				Mailbox:  "name20@ezbuy.com",
				Password: "pwd20",
			}

			pk, err := UserDBMgr(MySQL()).FindOne(unique)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(pk).ShouldNot(BeNil())
			obj, err := UserDBMgr(MySQL()).Fetch(pk)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(obj).ShouldNot(BeNil())
			Ω(obj.Mailbox).To(Equal(unique.Mailbox))
		})

		It("index", func() {
			sexIdx := &SexOfUserIDX{
				Sex: false,
			}
			total, pks, err := UserDBMgr(MySQL()).Find(sexIdx)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(int(total)).To(Equal(50))
			Ω(len(pks)).To(Equal(50))

			keys := make([]int32, len(pks))
			for i, pk := range pks {
				v := pk.(*IdOfUserPK)
				keys[i] = v.Id
			}

			objs, err := UserDBMgr(MySQL()).FetchByPrimaryKeys(keys)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(len(objs)).To(Equal(50))

		})

		It("range", func() {
			scope := &AgeOfUserRNG{
				AgeBegin: 10,
				AgeEnd:   35,
			}
			scope.Limit(10)
			total, pks, err := UserDBMgr(MySQL()).Range(scope)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(int(total)).To(Equal(24))
			Ω(len(pks)).To(Equal(10))

			keys := make([]int32, len(pks))
			for i, pk := range pks {
				v := pk.(*IdOfUserPK)
				keys[i] = v.Id
			}

			objs, err := UserDBMgr(MySQL()).FetchByPrimaryKeys(keys)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(len(objs)).To(Equal(10))
		})

		It("range.revert", func() {
			scope := &AgeOfUserRNG{}
			_, us, err := UserDBMgr(MySQL()).RangeRevert(scope)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(len(us)).To(Equal(100))
			// Ω(us[1].(int32) > us[0].(int32)).To(Equal(false))
		})

		It("fetch", func() {
			scope := &IdOfUserRNG{}
			_, us, err := UserDBMgr(MySQL()).Range(scope)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(len(us)).To(Equal(100))
			_, objs2, err := UserDBMgr(MySQL()).RangeFetch(scope)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(len(objs2)).To(Equal(100))
			for i, obj := range objs2 {
				Ω(obj.Name).To(Equal(fmt.Sprintf("name%d", i)))
			}
		})

		It("search", func() {
			us, err := UserDBMgr(MySQL()).Search("where age < 50 and sex = 1", "", "")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(len(us)).To(Equal(25))

			conditions := []string{}
			in := orm.NewFieldIN("age")
			in.Add(1)
			in.Add(3)
			in.Add(5)
			in.Add(7)
			in.Add(9)
			conditions = append(conditions, in.SQLFormat())
			ins, err := UserDBMgr(MySQL()).SearchConditions(conditions, "", 0, -1, in.SQLParams()...)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(len(ins)).To(Equal(5))

			cnt, err := UserDBMgr(MySQL()).SearchCount("where age < 50 and sex = 1")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(cnt).To(Equal(int64(25)))
		})

		It("query", func() {
			us, err := UserInfoDBMgr(MySQL()).QueryBySQL("SELECT `id`,`name`,`mailbox`, `password`, `sex` FROM users")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(len(us)).To(Equal(100))
		})

		Measure("mysql.bench", func(b Benchmarker) {
			b.Time("unique.runtime", func() {
				unique := &MailboxPasswordOfUserUK{
					Mailbox:  "name20@ezbuy.com",
					Password: "pwd20",
				}
				obj, err := UserDBMgr(MySQL()).FindOne(unique)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(obj).ShouldNot(BeNil())
			})
			b.Time("index.runtime", func() {
				sexIdx := &SexOfUserIDX{
					Sex: false,
				}
				_, us, err := UserDBMgr(MySQL()).Find(sexIdx)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(len(us)).To(Equal(50))
			})
			b.Time("range.runtime", func() {
				scope := &AgeOfUserRNG{
					AgeBegin: 10,
					AgeEnd:   35,
				}
				total, us, err := UserDBMgr(MySQL()).Range(scope)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(len(us)).To(Equal(int(total)))
			})
			b.Time("range.revert.runtime", func() {
				scope := &AgeOfUserRNG{}
				_, us, err := UserDBMgr(MySQL()).RangeRevert(scope)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(len(us)).To(Equal(100))
			})
			b.Time("fetch.runtime", func() {
				scope := &IdOfUserRNG{}
				_, us, err := UserDBMgr(MySQL()).Range(scope)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(len(us)).To(Equal(100))
				_, objs, err := UserDBMgr(MySQL()).RangeFetch(scope)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(len(objs)).To(Equal(100))
			})
		}, 1)
	})

})

var _ = Describe("redis-orm.redis", func() {

	MySQLSetup(&MySQLConfig{
		Host:     "localhost",
		Port:     3306,
		UserName: "ezorm_user",
		Password: "ezorm_pass",
		Database: "ezorm",
	})

	RedisSetUp(&RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Password: "",
	})

	BeforeEach(func() {
		tx, err := MySQL().BeginTx()
		Ω(err).ShouldNot(HaveOccurred())
		defer tx.Close()

		mgr := UserDBMgr(tx)

		users := []*User{}
		for i := 0; i < 100; i++ {
			user := UserMgr.NewUser()
			user.Name = fmt.Sprintf("name%d", i)
			user.Mailbox = fmt.Sprintf("name%d@ezbuy.com", i)
			user.HeadUrl = fmt.Sprintf("name%d.png", i)
			user.Password = fmt.Sprintf("pwd%d", i)
			if i%2 == 0 {
				user.Sex = true
			} else {
				user.Sex = false
			}
			user.Age = int32(i)
			user.CreatedAt = time.Now()
			user.UpdatedAt = user.CreatedAt
			user.Longitude = 103.754
			user.Latitude = 1.3282
			users = append(users, user)
		}

		n, err := mgr.BatchCreate(users)
		Ω(n).To(Equal(int64(100)))
		Ω(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		tx, err := MySQL().BeginTx()
		Ω(err).ShouldNot(HaveOccurred())
		defer tx.Close()

		mgr := UserDBMgr(tx)
		_, err = mgr.DeleteBySQL("")
		Ω(err).ShouldNot(HaveOccurred())
	})

	Describe("load", func() {
		It("mysql => redis", func() {
			Ω(UserRedisMgr(Redis()).Load(UserDBMgr(MySQL()))).ShouldNot(HaveOccurred())
			Ω(UserRedisMgr(Redis()).Clear()).ShouldNot(HaveOccurred())
			Ω(UserRedisMgr(Redis()).Load(UserDBMgr(MySQL()))).ShouldNot(HaveOccurred())
		})
	})

	Describe("crud", func() {
		var user, userWithExpire *User
		It("create", func() {
			user = UserMgr.NewUser()
			user.Id = 101
			user.Name = fmt.Sprintf("name%d", 101)
			user.Mailbox = fmt.Sprintf("name%d@ezbuy.com", 101)
			user.HeadUrl = fmt.Sprintf("name%d.png", 101)
			user.Password = fmt.Sprintf("pwd%d", 101)
			user.Sex = true
			user.Age = int32(32)
			user.CreatedAt = time.Now()
			user.UpdatedAt = user.CreatedAt
			user.Longitude = 103.754
			user.Latitude = 1.3282
			Ω(UserRedisMgr(Redis()).Create(user)).ShouldNot(HaveOccurred())

			obj, err := UserRedisMgr(Redis()).Fetch(user.GetPrimaryKey())
			Ω(err).ShouldNot(HaveOccurred())
			Ω(obj.Name).To(Equal(fmt.Sprintf("name%d", 101)))
		})
		It("createWithExpire", func() {
			userWithExpire = UserMgr.NewUser()
			userWithExpire.Id = 102
			userWithExpire.Name = fmt.Sprintf("name%d", 102)
			userWithExpire.Mailbox = fmt.Sprintf("name%d@ezbuy.com", 102)
			userWithExpire.HeadUrl = fmt.Sprintf("name%d.png", 102)
			userWithExpire.Password = fmt.Sprintf("pwd%d", 102)
			userWithExpire.Sex = true
			userWithExpire.Age = int32(33)
			userWithExpire.CreatedAt = time.Now()
			userWithExpire.UpdatedAt = user.CreatedAt
			userWithExpire.Longitude = 103.755
			userWithExpire.Latitude = 1.3283
			Ω(UserRedisMgr(Redis()).CreateWithExpire(userWithExpire, time.Second)).ShouldNot(HaveOccurred())

			obj, err := UserRedisMgr(Redis()).Fetch(userWithExpire.GetPrimaryKey())
			Ω(err).ShouldNot(HaveOccurred())
			Ω(obj.Name).To(Equal(fmt.Sprintf("name%d", 102)))

			time.Sleep(time.Second)
			_, err = UserRedisMgr(Redis()).Fetch(userWithExpire.GetPrimaryKey())
			fmt.Printf("createWithExpire after expire:%v", err)
			Ω(strings.Contains(err.Error(), "not exist")).Should(Equal(true))
		})
		It("update", func() {
			user.Age = int32(40)
			Ω(UserRedisMgr(Redis()).Update(user)).ShouldNot(HaveOccurred())
			obj, err := UserRedisMgr(Redis()).Fetch(user.GetPrimaryKey())
			Ω(err).ShouldNot(HaveOccurred())
			Ω(obj.Age).To(Equal(int32(40)))
		})
		It("updateWithExpire", func() {
			userWithExpire = UserMgr.NewUser()
			userWithExpire.Id = 102
			userWithExpire.Name = fmt.Sprintf("name%d", 102)
			userWithExpire.Mailbox = fmt.Sprintf("name%d@ezbuy.com", 102)
			userWithExpire.HeadUrl = fmt.Sprintf("name%d.png", 102)
			userWithExpire.Password = fmt.Sprintf("pwd%d", 102)
			userWithExpire.Sex = true
			userWithExpire.Age = int32(33)
			userWithExpire.CreatedAt = time.Now()
			userWithExpire.UpdatedAt = user.CreatedAt
			userWithExpire.Longitude = 103.755
			userWithExpire.Latitude = 1.3283
			Ω(UserRedisMgr(Redis()).Create(userWithExpire)).ShouldNot(HaveOccurred())

			obj, err := UserRedisMgr(Redis()).Fetch(userWithExpire.GetPrimaryKey())
			Ω(err).ShouldNot(HaveOccurred())
			Ω(obj.Name).To(Equal(fmt.Sprintf("name%d", 102)))

			userWithExpire.Age = int32(40)
			Ω(UserRedisMgr(Redis()).UpdateWithExpire(userWithExpire, time.Second)).ShouldNot(HaveOccurred())
			obj2, err := UserRedisMgr(Redis()).Fetch(userWithExpire.GetPrimaryKey())
			Ω(err).ShouldNot(HaveOccurred())
			Ω(obj2.Age).To(Equal(int32(40)))

			time.Sleep(time.Second)
			_, err = UserRedisMgr(Redis()).Fetch(userWithExpire.GetPrimaryKey())
			fmt.Printf("updateWithExpire after expire:%v", err)
			Ω(strings.Contains(err.Error(), "not exist")).Should(Equal(true))
		})
		It("delete", func() {
			Ω(UserRedisMgr(Redis()).Delete(user)).ShouldNot(HaveOccurred())
			_, err := UserRedisMgr(Redis()).Fetch(user.GetPrimaryKey())
			Ω(err).Should(HaveOccurred())
		})
		It("delete2", func() {
			Ω(UserRedisMgr(Redis()).Delete(userWithExpire)).ShouldNot(HaveOccurred())
			_, err := UserRedisMgr(Redis()).Fetch(userWithExpire.GetPrimaryKey())
			Ω(err).Should(HaveOccurred())
		})

		Measure("redis.bench", func(b Benchmarker) {
			b.Time("crud.runtime", func() {
				user = UserMgr.NewUser()
				user.Id = 101
				user.Name = fmt.Sprintf("name%d", 101)
				user.Mailbox = fmt.Sprintf("name%d@ezbuy.com", 101)
				user.HeadUrl = fmt.Sprintf("name%d.png", 101)
				user.Password = fmt.Sprintf("pwd%d", 101)
				user.Sex = true
				user.Age = int32(32)
				user.CreatedAt = time.Now()
				user.UpdatedAt = user.CreatedAt
				user.Longitude = 103.754
				user.Latitude = 1.3282
				Ω(UserRedisMgr(Redis()).Create(user)).ShouldNot(HaveOccurred())

				obj, err := UserRedisMgr(Redis()).Fetch(user.GetPrimaryKey())
				Ω(err).ShouldNot(HaveOccurred())
				Ω(obj.Name).To(Equal(fmt.Sprintf("name%d", 101)))

				user.Age = int32(40)
				Ω(UserRedisMgr(Redis()).Update(user)).ShouldNot(HaveOccurred())
				obj, err = UserRedisMgr(Redis()).Fetch(user.GetPrimaryKey())
				Ω(err).ShouldNot(HaveOccurred())
				Ω(obj.Age).To(Equal(int32(40)))

				Ω(UserRedisMgr(Redis()).Delete(user)).ShouldNot(HaveOccurred())
				_, err = UserRedisMgr(Redis()).Fetch(user.GetPrimaryKey())
				Ω(err).Should(HaveOccurred())
			})
		}, 1)
	})

	Describe("finder", func() {
		It("unique", func() {
			unique := &MailboxPasswordOfUserUK{
				Mailbox:  "name20@ezbuy.com",
				Password: "pwd20",
			}
			pk, err := UserRedisMgr(Redis()).FindOne(unique)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(pk).ShouldNot(BeNil())
			log.Println("redis.unique find pk =>", pk)

			usr, err := UserRedisMgr(Redis()).Fetch(pk)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(usr).ShouldNot(BeNil())
			Ω(usr.Mailbox).To(Equal("name20@ezbuy.com"))
		})
		It("index", func() {
			sexIdx := &SexOfUserIDX{
				Sex: false,
			}
			_, us, err := UserRedisMgr(Redis()).Find(sexIdx)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(len(us)).To(Equal(50))
		})
		It("range", func() {
			scope := &AgeOfUserRNG{
				AgeBegin: 10,
				AgeEnd:   35,
			}
			total, us, err := UserRedisMgr(Redis()).Range(scope)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(len(us)).To(Equal(int(total)))
		})
		It("range.revert", func() {
			scope := &AgeOfUserRNG{}
			_, us, err := UserRedisMgr(Redis()).RangeRevert(scope)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(len(us)).To(Equal(100))
			// Ω(us[1].(int32) > us[0].(int32)).To(Equal(false))
		})

		It("fetch", func() {
			scope := &IdOfUserRNG{}
			_, us, err := UserRedisMgr(Redis()).Range(scope)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(len(us)).To(Equal(100))
			_, objs, err := UserRedisMgr(Redis()).RangeFetch(scope)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(len(objs)).To(Equal(100))
			for i, obj := range objs {
				Ω(obj.Name).To(Equal(fmt.Sprintf("name%d", i)))
			}
			_, objs2, err := UserRedisMgr(Redis()).RangeFetch(scope)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(len(objs2)).To(Equal(100))
			for i, obj := range objs2 {
				Ω(obj.Name).To(Equal(fmt.Sprintf("name%d", i)))
			}
		})

		Measure("redis.bench", func(b Benchmarker) {
			b.Time("unique.runtime", func() {
				unique := &MailboxPasswordOfUserUK{
					Mailbox:  "name20@ezbuy.com",
					Password: "pwd20",
				}
				obj, err := UserRedisMgr(Redis()).FindOne(unique)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(obj).ShouldNot(BeNil())
			})
			b.Time("index.runtime", func() {
				sexIdx := &SexOfUserIDX{
					Sex: false,
				}
				_, us, err := UserRedisMgr(Redis()).Find(sexIdx)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(len(us)).To(Equal(50))
			})
			b.Time("range.runtime", func() {
				scope := &AgeOfUserRNG{
					AgeBegin: 10,
					AgeEnd:   35,
				}
				_, us, err := UserRedisMgr(Redis()).Range(scope)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(len(us)).To(Equal(24))
				// Ω(us[1].(int32) > us[0].(int32)).To(Equal(true))
			})
			b.Time("range.revert.runtime", func() {
				scope := &AgeOfUserRNG{}
				_, us, err := UserRedisMgr(Redis()).RangeRevert(scope)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(len(us)).To(Equal(100))
				// Ω(us[1].(int32) > us[0].(int32)).To(Equal(false))
			})
			b.Time("fetch.runtime", func() {
				scope := &AgeOfUserRNG{}
				_, us, err := UserRedisMgr(Redis()).RangeRevert(scope)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(len(us)).To(Equal(100))
				_, objs, err := UserRedisMgr(Redis()).RangeRevertFetch(scope)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(len(objs)).To(Equal(100))
			})
		}, 1)
	})
})
