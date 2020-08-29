package lytoken

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

// 羚羊云 Token 控制位选项功能列表
const (
	// 从左到右, 高位字节开始

	// byte 1, 直播及安全控制, 0 ~ 7, 0 ~ 5 used, 6 ~ 7 reserved
	CtlRTMPLive      = "rtmp_live"      // 0: RTMP 直播
	CtlHLSLive       = "hls_live"       // 1: HLS 直播
	CtlValidateIP    = "validate_ip"    // 2: 验证 IP 段
	CtlValidateRefer = "validate_refer" // 3: 验证 Refer 段
	CtlAcceptUDP     = "accept_udp"     // 4: 接受 UDP 请求
	CtlUploadObject  = "upload_object"  // 5: 上传对象

	// byte2, 存储控制, 8 ~ 15
	CtlStorage             = "storage"               // 8 ~ 11: 存储控制, 具体值为 0 ~ 15, 4 bits
	CtlUseCircleStorage    = "use_circle_storage"    // 12: 使用循环存储
	CtlUseEventStorage     = "use_event_storage"     // 13: 使用事件存储
	CtlAllowObjectDownload = "allow_object_download" // 14: 允许下载对象
	CtlValidateObjectID    = "validate_object_id"    // 15: 验证 Object ID

	// byte3, 播放控制, 16 ~ 23
	CtlWatchPublic    = "watch_public"     // 16: 观看公众设备视频
	CtlWatchPrivate   = "watch_private"    // 17: 观看私有设备视频
	CtlWatchTimeShift = "watch_time_shift" // 18: 观看时移
	CtlWatchVideo     = "watch_video"      // 19: 观看录像
	CtlSoundPassBack  = "sound_pass_back"  // 20: 语音回传
	CtlVideoPassBack  = "video_pass_back"  // 21: 视频回传
	CtlGetSnapshot    = "get_snapshot"     // 22: 观看封面截图
	CtlSound          = "sound"            // 23: 收听声音

	// byte4, 24 ~ 31 reserved
)

const (
	// defaultControlValue 功能选项默认值
	defaultControlValue = "0"
)

// 控制位各个功能选项值, 默认不开启
var (
	// 直播及安全控制
	OptionRTMPLive = CtlOption{
		Name:     CtlRTMPLive,
		position: 0,
		len:      1,
		value:    "1",
	}
	OptionHLSLive = CtlOption{
		Name:     CtlHLSLive,
		position: 1,
		len:      1,
		value:    "1",
	}
	OptionValidateIP = CtlOption{
		Name:     CtlValidateIP,
		position: 2,
		len:      1,
		value:    defaultControlValue,
	}
	OptionValidateRefer = CtlOption{
		Name:     CtlValidateRefer,
		position: 3,
		len:      1,
		value:    defaultControlValue,
	}
	OptionAcceptUDP = CtlOption{
		Name:     CtlAcceptUDP,
		position: 4,
		len:      1,
		value:    defaultControlValue,
	}
	OptionUploadObject = CtlOption{
		Name:     CtlUploadObject,
		position: 5,
		len:      1,
		value:    defaultControlValue,
	}

	// 存储控制
	OptionStorage = CtlOption{
		Name:     CtlStorage,
		position: 8,
		len:      4,
		value:    "0000",
	}
	OptionUseCircleStorage = CtlOption{
		Name:     CtlUseCircleStorage,
		position: 12,
		len:      1,
		value:    "1",
	}
	OptionUseEventStorage = CtlOption{
		Name:     CtlUseEventStorage,
		position: 13,
		len:      1,
		value:    "1",
	}
	OptionAllowObjectDownload = CtlOption{
		Name:     CtlAllowObjectDownload,
		position: 14,
		len:      1,
		value:    defaultControlValue,
	}
	OptionValidateObjectID = CtlOption{
		Name:     CtlValidateObjectID,
		position: 15,
		len:      1,
		value:    defaultControlValue,
	}

	// 播放控制
	OptionWatchPublic = CtlOption{
		Name:     CtlWatchPublic,
		position: 16,
		len:      1,
		value:    defaultControlValue,
	}
	OptionWatchPrivate = CtlOption{
		Name:     CtlWatchPrivate,
		position: 17,
		len:      1,
		value:    defaultControlValue,
	}
	OptionWatchTimeShift = CtlOption{
		Name:     CtlWatchTimeShift,
		position: 18,
		len:      1,
		value:    defaultControlValue,
	}
	OptionWatchVideo = CtlOption{
		Name:     CtlWatchVideo,
		position: 19,
		len:      1,
		value:    defaultControlValue,
	}
	OptionSoundPassBack = CtlOption{
		Name:     CtlSoundPassBack,
		position: 20,
		len:      1,
		value:    defaultControlValue,
	}
	OptionVideoPassBack = CtlOption{
		Name:     CtlVideoPassBack,
		position: 21,
		len:      1,
		value:    defaultControlValue,
	}
	OptionGetSnapshot = CtlOption{
		Name:     CtlGetSnapshot,
		position: 22,
		len:      1,
		value:    defaultControlValue,
	}
	OptionSound = CtlOption{
		Name:     CtlSound,
		position: 23,
		len:      1,
		value:    defaultControlValue,
	}

	// OptionList 所有使用的选项
	OptionList = []CtlOption{
		OptionRTMPLive, OptionHLSLive, OptionValidateIP, OptionValidateRefer,
		OptionAcceptUDP, OptionUploadObject, OptionStorage, OptionUseCircleStorage,
		OptionUseEventStorage, OptionAllowObjectDownload, OptionValidateObjectID,
		OptionWatchPublic, OptionWatchPrivate, OptionWatchTimeShift, OptionWatchVideo,
		OptionSoundPassBack, OptionVideoPassBack, OptionGetSnapshot, OptionSound,
	}
	// StorageOptions 录像存储相关
	StorageOptions = map[string]CtlOption{}

	// Options 全局的控制段选项管理
	Options *ControlOptions
)

// ControlOptions 羚羊云 Token 控制字段功能选项
type ControlOptions struct {
	optionsNameMap     map[string]CtlOption
	optionsPositionMap map[int]CtlOption
}

// OptionByName 根据名称获取选项
func (options *ControlOptions) OptionByName(name string) (CtlOption, bool) {
	option, ok := options.optionsNameMap[name]
	if !ok {
		return CtlOption{}, false
	}
	return option.Copy(), true
}

// OptionByPosition 根据位置获取选项
func (options *ControlOptions) OptionByPosition(position int) (CtlOption, bool) {
	option, ok := options.optionsPositionMap[position]
	if !ok {
		return CtlOption{}, false
	}
	return option.Copy(), true
}

// StorageOptionByExpireType 根据 expire type 值返回存储选项
func (options *ControlOptions) StorageOptionByExpireType(expireType int) (CtlOption, bool) {
	if expireType >= 16 || expireType < 0 {
		return CtlOption{}, false
	}

	opLen := 4
	//十进制转换二进制
	binStr := strconv.FormatInt(int64(expireType), 2)
	length := len(binStr)
	if length < opLen {
		binStr = fmt.Sprintf("%v%v", strings.Repeat("0", opLen-length), binStr)
	}
	return OptionStorage.Value(binStr), true
}

// Control 羚羊云 Token 控制数据
type Control struct {
	bytes   []byte
	options map[string]CtlOption
}

// NewControl 创建一个新的控制变量
func NewControl() *Control {
	binStr := strings.Repeat("0", 32)
	return &Control{
		bytes:   []byte(binStr),
		options: make(map[string]CtlOption),
	}
}

// ControlNumber 从数值创建一个新的控制变量
func ControlNumber(number uint32) *Control {
	binStr := strconv.FormatUint(uint64(number), 2)
	if len(binStr) < 32 {
		lackCnt := 32 - len(binStr)
		fillZeroStr := strings.Repeat("0", lackCnt)
		binStr = fmt.Sprintf("%v%v", fillZeroStr, binStr)
	}

	options := map[string]CtlOption{}
	for i := range binStr {
		option, ok := Options.OptionByPosition(i)
		if !ok {
			continue
		}
		option.value = binStr[option.position : option.position+option.len]
		options[option.Name] = option
	}
	return &Control{
		bytes:   []byte(binStr),
		options: options,
	}
}

// Number 获取控制数值
func (ctl *Control) Number() uint32 {
	for _, option := range ctl.options {
		for i, v := range option.value {
			ctl.bytes[option.position+i] = byte(v)
		}
	}
	binStr := string(ctl.bytes)
	num, err := strconv.ParseUint(binStr, 2, 32)
	if err != nil {
		log.Printf("lytoken: invalid binary string %v: %v\n", binStr, err)
		return 0
	}
	return uint32(num)
}

// OptionValue 获取指定名称选项的值
func (ctl *Control) OptionValue(name string) string {
	var option CtlOption
	option, ok := ctl.options[name]
	if !ok {
		option, ok = Options.OptionByName(name)
	}
	if ok {
		return string(ctl.bytes[option.position : option.position+option.len])
	}
	return ""
}

// SetOption 设置控制项
func (ctl *Control) SetOption(option CtlOption, value string) *Control {
	option.SetValue(value)
	ctl.options[option.Name] = option
	ctl.setOptionToBytes(option)
	return ctl
}

// SetStorageByExpireType 根据 expireType 值设置存储选项, 值 0 ~ 15, 超范围使用 0
func (ctl *Control) SetStorageByExpireType(expireType int) *Control {
	option, ok := Options.StorageOptionByExpireType(expireType)
	if !ok {
		option = OptionStorage.Value("0000")
	}
	ctl.options[option.Name] = option
	ctl.setOptionToBytes(option)
	return ctl
}

// setOptionToBytes 设置选项值到 bytes 中
func (ctl *Control) setOptionToBytes(option CtlOption) *Control {
	for i, v := range option.value {
		ctl.bytes[option.position+i] = byte(v)
	}
	return ctl
}

// CtlOption 羚羊 Token 控制数据选项
type CtlOption struct {
	Name string

	position int
	len      int
	value    string
}

// SetValue 设置控制选项的值
func (option *CtlOption) SetValue(value string) *CtlOption {
	if !validCtlOptionValue(value) {
		return option
	}
	if len(value) != option.len {
		return option
	}
	option.value = value
	return option
}

// Copy 从现有的选项拷贝一个新的选项
func (option CtlOption) Copy() CtlOption {
	return CtlOption{
		Name:     option.Name,
		position: option.position,
		len:      option.len,
		value:    option.value,
	}
}

// Value 从现有选项拷贝一个新选项并设置值
func (option CtlOption) Value(value string) CtlOption {
	newOption := option.Copy()
	(&newOption).SetValue(value)
	return newOption
}

// validCtlOptionValue 检查是否为有效的控制取值, 必须为 0 或 1 字符
func validCtlOptionValue(value string) bool {
	values := map[string]bool{
		"0": true,
		"1": true,
	}
	for _, a := range value {
		if _, ok := values[string(a)]; !ok {
			return false
		}
	}
	return true
}

func initControlOptions() {
	optionsNameMap := map[string]CtlOption{}
	optionsPositionMap := map[int]CtlOption{}
	for _, option := range OptionList {
		optionsNameMap[option.Name] = option
		optionsPositionMap[option.position] = option
	}

	Options = &ControlOptions{
		optionsNameMap:     optionsNameMap,
		optionsPositionMap: optionsPositionMap,
	}
}
