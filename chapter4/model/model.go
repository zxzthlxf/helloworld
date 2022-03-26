package model

type ExecTime struct {
	StartTime int64 `bson:"startTime"` //开始时间
	EndTime   int64 `bson:"endTime"`   //结束时间
}

type LogRecord struct {
	JobName string   `bson:"jobName"` //任务名
	Command string   `bson:"command"` //shell命令
	Err     string   `bson:"err"`     //脚本错误
	Content string   `bson:"content"` //脚本输出
	Tp      ExecTime //执行时间
}

type FindByJobName struct {
	JobName string `bson:"jobName"` //任务名
}

type UpdateByJobName struct {
	Command string `bson:"command"` //shell命令
	Content string `bson:"content"` //脚本输出
}

type DeleteCond struct {
	BeforeCond TimeBeforeCond `bson"tp.startTime"`
}

type TimeBeforeCond struct {
	BeforeTime int64 `bson:"$lt"`
}

type ExecTimeFilter struct {
	StartTime interface{} `bson:"tp.startTime,omitempty"` //开始时间
	EndTime   interface{} `bson:"tp.endTime,omitempty"`   //结束时间
}

type LogRecordFilter struct {
	ID      interface{} `bson:"_id,omitempty"`
	JobName interface{} `bson:"jobName,omitempty" json:"jobName"` //任务名
	Command interface{} `bson:"command,omitempty"`                //shell命令
	Err     interface{} `bson:"err,omitempty"`                    //脚本错误
	Content interface{} `bson:"tp,omitempty"`                     //执行时间
}

//小于示例
type Lt struct {
	Lt int64 `bson:"$lt"`
}

//分组示例
type Group struct {
	Group interface{} `bson:"$group"`
}

//求和示例
type Sum struct {
	Sum interface{} `bson:"$sum"`
}
