package main

import (
	"fmt"
	//"fmt"
	// orm "github.com/astaxie/beego/client/orm"
	orm "github.com/eopenio/borm"
	_ "github.com/go-sql-driver/mysql"
	"helloworld/chapter4/model"
	"io"
	//"io"
)

func init() {
	orm.RegisterDriver("mysql", orm.DRMySQL)
	orm.RegisterDataBase("default", "mysql", "root:Jzyz.8888@tcp(192.168.8.151:3306)/chapter4?charset=utf8mb4")
	orm.RegisterModel(new(model.BeegoUser))
}

func main() {
	/*
		//插入数据
		o:=orm.NewOrm()
		user:=new(model.BeegoUser)
		user.Name="jim"
		user.Phone="13600000000"
		fmt.Println(o.Insert(user))
	*/
	//查询数据

	o := orm.NewOrm()
	user := model.BeegoUser{}
	user.Id = 1
	err := o.Read(&user)

	if err == orm.ErrNoRows {
		fmt.Println("查询不到")
	} else if err == orm.ErrMissPK {
		fmt.Println("找不到主键")
	} else {
		fmt.Println(user.Id, user.Name, user.Phone)
	}
	/*
			//更新数据
			o:=orm.NewOrm()
			user:=model.BeegoUser{}
			user.Id=1
			user.Name="James"

			num,err:=o.Update(&user)
			if err!=nil{
				fmt.Println("更新失败")
			}else{
				fmt.Println("更新数据影响的行数：",num)
			}


			//删除数据
			o:=orm.NewOrm()
			user:=model.BeegoUser{}
			user.Id=3

			if num,err:=o.Delete(&user);err!=nil{
				fmt.Println("删除失败")
			}else{
				fmt.Println("删除数据影响的行数：",num)
			}

		//原生sql查询
		o := orm.NewOrm()
		var r orm.RawSeter
		r,err := o.Raw("update beego_user set name=? where name=?", "jack", "jim")
		if err!=nil{
			fmt.Println(err)
		}

			//事务处理
			o:=orm.NewOrm()
			o.Begin()
			user1:=model.BeegoUser{}
			user1.Id=6
			user1.Name="james"

			user2:=model.BeegoUser{}
			user2.Id=12
			user2.Name="Wade"

			_,err1:=o.Update(&user1)
			_,err2:=o.Insert(&user2)
			if err1!=nil||err2!=nil{
				o.Rollback()
			}else{
				o.Commit()
			}*/

	//在调试模式下打印查询语句
	orm.Debug = true
	var w io.Writer
	orm.DebugLog = orm.NewLog(w)

}
