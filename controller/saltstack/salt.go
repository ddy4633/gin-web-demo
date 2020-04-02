package saltstack

import (
	"bytes"
	"encoding/json"
	"gin-web-demo/conf"
	"gin-web-demo/dao"
	"gin-web-demo/tools"
	"io/ioutil"
	"net/http"
	"time"
)

//salt控制器
type SaltController struct {
}

//获取salt初始化信息
func (s *SaltController) GetToken() (saltinfo conf.Returninfo) {
	/*
		如果是带有HTTPs的则还需要传递TLS进Client中
	*/

	//配置请求信息
	info := &conf.Info{
		Username: "saltapi",
		Password: "saltapi",
		Eauth:    "pam",
	}
	//序列化
	buf, err := json.Marshal(info)
	if !tools.CheckERR(err, "Json Marshal is Failed") {
		return saltinfo
	}
	//新建一个请求
	re, err := http.NewRequest("POST", conf.Config.Conf.URL_LOGIN, bytes.NewBuffer(buf))
	if !tools.CheckERR(err, "Creata New Request") {
		return saltinfo
	}
	//设置请求格式
	re.Header.Set("Accept", conf.Json_Accept)
	re.Header.Set("Content-Type", conf.Json_Content_Type)
	//新建一个请求
	client := http.Client{
		Timeout: 3 * time.Second,
	}
	//创建请求
	respon, err := client.Do(re)
	if !tools.CheckERR(err, "Create Client Request") {
		return saltinfo
	}
	defer respon.Body.Close()
	//读返回信息
	body, err := ioutil.ReadAll(respon.Body)
	if !tools.CheckERR(err, "ReadALL response Body Failed") {
		return saltinfo
	}
	//反序列化
	err = json.Unmarshal(body, &saltinfo)
	if !tools.CheckERR(err, "Json Unmarshal Returninfo Failed") {
		return saltinfo
	}
	//fmt.Println(saltinfo)
	return saltinfo
}

//执行指定的模块
func (s *SaltController) PostModulJob(token string, cmd *conf.JobRunner) *conf.JobReturn {
	var (
		//临时使用
		relist conf.JobReturn
	)
	//调用构造函数
	response := pulicPost(token, cmd)
	//读信息
	infodata, _ := ioutil.ReadAll(response.Body)
	json.Unmarshal(infodata, &relist)
	//fmt.Println("infodata=", infodata)
	return &relist
}

//公共的POST整理
func pulicPost(token string, para *conf.JobRunner) (response *http.Response) {
	//构建json参数
	cmd := &conf.JobRunner{
		Client:    para.Client,
		Tgt:       para.Tgt,
		Fun:       para.Fun,
		Arg:       para.Arg,
		Expr_form: para.Expr_form,
	}
	//Object序列化
	data, err := json.Marshal(cmd)
	tools.CheckERR(err, "PostModulJob Object json marshal Is Failed")
	//新建请求
	re, err := http.NewRequest("POST", conf.Config.Conf.URL, bytes.NewBuffer(data))
	if !tools.CheckERR(err, "Create PostModulJob Request Failed") {
		return response
	}
	defer re.Body.Close()
	//设置请求头
	re.Header.Set("Accept", conf.Json_Accept)
	re.Header.Set("X-Auth-Token", token)
	re.Header.Set("Content-Type", conf.Json_Content_Type)
	//fmt.Println(re)
	//新建Client
	client := http.Client{}
	//请求对端
	response, err = client.Do(re)
	if !tools.CheckERR(err, "PostModulJob Client Request is Failed") {
		return
	}
	return response
}

//执行Job任务查询
func (s *SaltController) QueryJob(jobid string, token string) conf.JobInfo {
	var (
		buf    []byte
		result conf.JobInfo
	)
	//新建请求
	re, err := http.NewRequest("GET", conf.Config.Conf.URL_JOBS+"/"+jobid, bytes.NewBuffer(buf))
	if !tools.CheckERR(err, "Create PostModulJob Request Failed") {
		return result
	}
	defer re.Body.Close()
	//设置请求头
	re.Header.Set("Accept", conf.Json_Accept)
	re.Header.Set("X-Auth-Token", token)
	//re.Header.Set("Content-Type", conf.Json_Content_Type)
	//fmt.Println(re)
	//新建Client
	client := http.Client{}
	//请求对端
	response, err := client.Do(re)
	if !tools.CheckERR(err, "PostModulJob Client Request is Failed") {
		return result
	}
	//读信息
	infodata, _ := ioutil.ReadAll(response.Body)
	//反序列化
	json.Unmarshal(infodata, &result)
	if !tools.CheckERR(err, "JobResult Unmarshal is Failed") {
		return result
	}
	//fmt.Println("序列化后的数据", infodata)
	return result
}

//返回任务的最终执行结果
func (s *SaltController) ReturnResult(jid string) string {
	//获取数据源
	data := reddao.GetDate(jid)
	return data
}

//获取CMDB的认证Token
func (s *SaltController) GetCMDBAUTH() error {
	var obj conf.TokenCmdb
	//构建对象
	auth := &conf.AuthCmdb{
		UserName: conf.Config.Conf.Ldap_user,
		PassWord: tools.GetLdapPasswd(conf.Config.Conf.Ldap_passwd),
	}
	//序列化
	au, err := json.Marshal(auth)
	if !tools.CheckERR(err, "") {
		return err
	}
	//构建连接
	req, err := http.NewRequest("POST", conf.Config.Conf.CMDB_api, bytes.NewBuffer(au))
	tools.CheckERR(err, "New CMDB Request URL IS Failed")
	//设置request
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "JWT ")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	//请求连接等待返回
	client := http.Client{}
	repon, err := client.Do(req)
	//读信息
	infodata, _ := ioutil.ReadAll(repon.Body)
	tools.CheckERR(err, "Request CMDB IS Failed")
	//反序列化
	err = json.Unmarshal(infodata, &obj)
	tools.CheckERR(err, "json Unmarshal CMDB IS Failed")
	//存数据库
	err = dao.RedisHandle{}.InsertTTLData("AuthToken", obj.Token, "EX", "18000")
	tools.CheckERR(err, "json Unmarshal CMDB IS Failed")
	return err
}

//查询CMDB的接口
//func (s *SaltController) GetCMDBInfo(IP string) []string {
//	//构建参数
//	buf := []byte(IP)
//	//构建连接
//	req, err := http.NewRequest("POST", conf.Config.Conf.CMDB_api, bytes.NewBuffer(buf))
//	tools.CheckERR(err, "New CMDB Request URL IS Failed")
//	//设置request
//	req.Header.Set("Content-Type", "application/json")
//	req.Header.Set("Authorization", "JWT ")
//	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
//	//请求连接等待返回
//	client := http.Client{}
//	repon, err := client.Do(req)
//	tools.CheckERR(err, "Request CMDB IS Failed")
//	//ioutil.ReadAll(repon,)
//
//}
