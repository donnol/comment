package comment

// User 用户
// 这是一个表示用户的结构体
type User struct {
	*Model
	Basic                 // 基础信息
	Age         int       // 年纪
	AddressList []Address // 地址列表
}

// Address 地址
type Address struct {
	Basic           // 基础信息
	Position string // 位置
}

// Basic 基础信息
type Basic struct {
	Name string // 名字
}

// Model 模型
type Model struct {
	InnerModel
	Help func() // 帮助函数
}

// InnerModel 内部模型
type InnerModel struct {
	CornerStone string // 基石
}
