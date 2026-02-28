package server

import (
	"github.com/sirupsen/logrus"
)

// UserInfo 用户基本信息
type UserInfo struct {
	Subscribe      int32  `json:"subscribe"`       // 用户是否订阅该公众号标识，值为0时，代表此用户没有关注该公众号，拉取不到其余信息
	OpenID         string `json:"openid"`          // 用户的标识，对当前公众号唯一
	Language       string `json:"language"`        // 用户的语言，简体中文为zh_CN
	SubscribeTime  int32  `json:"subscribe_time"`  // 用户关注时间，为时间戳。如果用户曾多次关注，则取最后关注时间
	UnionID        string `json:"unionid"`         // 只有在用户将公众号绑定到微信开放平台帐号后，才会出现该字段
	Remark         string `json:"remark"`          // 公众号运营者对粉丝的备注，公众号运营者可在微信公众平台用户管理界面对粉丝添加备注
	GroupID        int32  `json:"groupid"`         // 用户所在的分组ID（兼容旧的用户分组接口）
	TagIDList      []int32 `json:"tagid_list"`     // 用户被打上的标签ID列表
	SubscribeScene string `json:"subscribe_scene"` // 返回用户关注的渠道来源
	QRScene        int    `json:"qr_scene"`        // 二维码扫码场景（开发者自定义）
	QRSceneStr     string `json:"qr_scene_str"`    // 二维码扫码场景描述（开发者自定义）
	Nickname       string `json:"nickname"`        // 用户的昵称
	Sex            int32  `json:"sex"`             // 用户的性别，值为1时是男性，值为2时是女性，值为0时是未知
	City           string `json:"city"`            // 用户所在城市
	Country        string `json:"country"`         // 用户所在国家
	Province       string `json:"province"`        // 用户所在省份
	Headimgurl     string `json:"headimgurl"`      // 用户头像，最后一个数值代表正方形头像大小（有0、46、64、96、132数值可选，0代表640*640正方形头像），用户没有头像时该项为空
}

// GetUserInfo 获取用户基本信息
// openid: 用户的标识，对当前公众号唯一
func (wChat *WeChat) GetUserInfo(openid string) (*UserInfo, error) {
	// 调用微信SDK获取用户信息
	user := wChat.Account.GetUser()
	info, err := user.GetUserInfo(openid)
	if err != nil {
		logrus.Errorf("get user info from wechat failed: %v", err)
		return nil, err
	}

	// 转换为我们的结构体
	userInfo := &UserInfo{
		Subscribe:      info.Subscribe,
		OpenID:         info.OpenID,
		Language:       info.Language,
		SubscribeTime:  info.SubscribeTime,
		UnionID:        info.UnionID,
		Remark:         info.Remark,
		GroupID:        info.GroupID,
		TagIDList:      info.TagIDList,
		SubscribeScene: info.SubscribeScene,
		QRScene:        info.QrScene,
		QRSceneStr:     info.QrSceneStr,
		Nickname:       info.Nickname,
		Sex:            info.Sex,
		City:           info.City,
		Country:        info.Country,
		Province:       info.Province,
		Headimgurl:     info.Headimgurl,
	}

	logrus.Infof("get user info success for openid: %s, nickname: %s", openid, userInfo.Nickname)
	return userInfo, nil
}
