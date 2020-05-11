package saltstack

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gin-web-demo/conf"
	"gin-web-demo/dao"
	"gin-web-demo/tools"
	"io/ioutil"
	"net/http"
	"reflect"
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
		Username: conf.Config.Conf.Saltauth[0].Username,
		Password: conf.Config.Conf.Saltauth[0].Password,
		Eauth:    conf.Config.Conf.Saltauth[0].Eauth,
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

//异步执行指定的模块
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

//同步执行模块
func (s *SaltController) PostRsyncModulJob(token string, cmd *conf.JobRunner) string {
	var (
		//临时使用
		data conf.CheckActive
	)
	//调用构造函数
	response := pulicPost(token, cmd)
	//读信息
	infodata, _ := ioutil.ReadAll(response.Body)
	json.Unmarshal(infodata, &data)
	fmt.Println("infodata=", infodata)
	//反射出结果
	obj := data.Return[0].(map[string]interface{})
	//返回对象
	result := obj[cmd.Tgt].(string)
	fmt.Println("infodata=", result)
	return result
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
	if !tools.CheckERR(err, "PostModulJob Object json marshal Is Failed") {
		return response
	}
	conf.WriteLog(fmt.Sprintf("%s[Return]cmd=%s,序列化后=%s\n", time.Now().Format("2006-01-02 15:04:05"), cmd, string(data)))
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
	conf.WriteLog(fmt.Sprintf("%s[Return]re.body=%s\n", time.Now().Format("2006-01-02 15:04:05"), re.Body))
	//fmt.Println(re,"conf.Config.Conf.URL=",conf.Config.Conf.URL)
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
	//请求连接等待返回
	client := http.Client{}
	repon, err := client.Do(req)
	//读信息
	infodata, _ := ioutil.ReadAll(repon.Body)
	tools.CheckERR(err, "Request CMDB IS Failed")
	//反序列化
	err = json.Unmarshal(infodata, &obj)
	tools.CheckERR(err, "json Unmarshal CMDB IS Failed")
	conf.WriteLog(fmt.Sprintf("%s[Return]AuthToken获取回来的消息为=%s\n", time.Now().Format("2006-01-02 15:04:05"), string(infodata)))
	//存数据库
	err = dao.RedisHandle{}.InsertTTLData("AuthToken", obj.Token, "EX", "18000")
	tools.CheckERR(err, "json Unmarshal CMDB IS Failed")
	return err
}

type Ip struct {
	IP string `json:"ip"`
}

//查询CMDB的接口
func (s *SaltController) GetCMDBInfo(ips string) (string, error) {
	var (
		retruninfo conf.Retuencmdb
		token      string
	)
	//取token信息
	if token = reddao.GetDate("AuthToken"); len(token) < 0 {
		return "", errors.New("Get CMDB AuthToken is Failed,Please check !")
	}
	//构建参数
	ip := &Ip{IP: ips}
	buf, err := json.Marshal(&ip)
	if !tools.CheckERR(err, "New CMDB Request URL IS Failed") {
		return "", errors.New("ip参数序列化失败请检查")
	}
	//构建连接
	req, err := http.NewRequest("POST", conf.Config.Conf.Cmdb_infoapi, bytes.NewBuffer(buf))
	tools.CheckERR(err, "New CMDB Request URL IS Failed")
	//设置request
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "JWT"+" "+token)
	//请求连接等待返回
	client := http.Client{}
	repon, err := client.Do(req)
	tools.CheckERR(err, "Request CMDB IS Failed")
	info, _ := ioutil.ReadAll(repon.Body)
	json.Unmarshal(info, &retruninfo)
	if retruninfo.Code != 00000 {
		return "", errors.New("Dont's Get CMDB minion Address,Please check request!")
	}
	return retruninfo.Data.IPgroup, nil
}

//salt-minion存活检测
func (s *SaltController) ActiveSalt(address string) (bool, string) {
	//获取token信息
	token, err := s.Check()
	fmt.Println("请求进来了", address, "", token, err)
	if !tools.CheckERR(err, "获取token失败") {
		return false, fmt.Sprintf("内部获取token失败,ERROR=%s", err)
	}
	fmt.Println("请求进来了", token)
	//构建json参数
	cmd := &conf.JobRunner{
		Client: "local",
		Tgt:    address,
		Fun:    "test.ping",
	}
	fmt.Printf("token=%s,cmd=%s\n", token, cmd)
	//请求对端
	obj := pulicPost(token, cmd)
	data, err := ioutil.ReadAll(obj.Body)
	if !tools.CheckERR(err, "read checkactive is Failed") {
		return false, fmt.Sprintf("读取ioutil失败,ERROR=%s", err)
	}
	check := &conf.CheckActive{}
	err = json.Unmarshal(data, check)
	tools.CheckERR(err, "ActiveCheck json unmarshal is failed!")
	conf.WriteLog(fmt.Sprintf("%s[Return]ActiveSalt返回信息为=%s\n", time.Now().Format("2006-01-02 15:04:05"), check))
	//防止越界
	//if reflect.ValueOf(check.Return).IsNil()||reflect.ValueOf(check.Return).IsValid(){
	//	return false,errors.New("发生未知错误,数组越界")
	//}
	//断言类型转换换
	checks, ok := check.Return[0].(map[string]interface{})
	//if !ok {
	//	return false,errors.New("发生未知错误,数组越界")
	//}
	//if len(checks) < 1 {
	//	fmt.Println("checks len is =", len(checks))
	//	return false, errors.New("该salt-minion不存在!")
	//}
	//if !checks[address].(bool) {
	//	//是否存活
	//	conf.WriteLog(fmt.Sprintf("%s[salt-check]存活检测失败状态为=%s\n", tools.GetTimeNow(), check))
	//	return false, errors.New("salt-minion死亡状态!请检查")
	//}
	//(只要满足以上3种情况其一)均为无效值
	switch {
	case len(checks) < 1:
		fmt.Println("checks len is =", len(checks))
		return false, "该salt-minion不存在!"
	case !checks[address].(bool):
		//是否存活
		conf.WriteLog(fmt.Sprintf("%s[salt-check]存活检测失败状态为=%s\n", tools.GetTimeNow(), check))
		return false, "salt-minion死亡状态!请检查"
	case !ok:
		return false, "发生未知错误,数组越界"
	case reflect.ValueOf(check.Return).IsNil() || reflect.ValueOf(check.Return).IsValid():
		return false, "发生未知错误,数组越界"
	}
	conf.WriteLog(fmt.Sprintf("%s[Return]判断结果的错误信息为Err=%s\n", time.Now().Format("2006-01-02 15:04:05"), err))
	return true, "salt-minion存活ping通畅!"
}

//salt-Token调用检测
func (s *SaltController) Check() (tokens string, err error) {
	//获取Token信息
	if tokens = reddao.GetDate("token"); tokens == "" {
		fmt.Println("请求执行到获取token了")
		tokens = s.GetToken().Return[0].Token
		err = reddao.InsertTTLData("token", tokens, "EX", "3600")
		if !tools.CheckERR(err, "Inserter Token Failed") {
			return
		}
		conf.WriteLog(fmt.Sprintf("[info]获取Token信息=%s\n", tokens))
	}
	fmt.Println("请求返回获取token了")
	return
}
