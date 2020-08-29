package lytoken

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

const (
	// Delimiter token 各个段之间的分隔符
	Delimiter = "_"
	// tokenSectionAmount Token 字符串分段数量
	tokenSectionAmount = 4
)

var (
	// defaultDuration 默认的有效时长
	defaultDuration = 24 * time.Hour
	// defaultControl 默认的控制位
	defaultControl = NewControl()
)

// 错误变量
var (
	ErrInvalidTokenStr = errors.New("lytoken: invalid token string")    // 无效的 Token 字符串
	ErrInvalidCID      = errors.New("lytoken: invalid cid section")     // 无效的 CID
	ErrInvalidExpired  = errors.New("lytoken: invalid expired section") // 无效的 Expired
	ErrInvalidControl  = errors.New("lytoken: invalid control section") // 无效的 Control
)

// Token 羚羊云 token
type Token struct {
	CID     uint32
	Ctl     *Control
	Expired uint32

	tokenStr string
}

// New 构造一个新的 token
func New(cid, ctlNum uint32, duration time.Duration) *Token {
	expired := time.Now().Add(duration).Unix()
	ctl := ControlNumber(ctlNum)
	return &Token{
		CID:     cid,
		Ctl:     ctl,
		Expired: uint32(expired),
	}
}

// NewDefault 构造一个新的 control 为 0, expired 为 24 小时的 token
func NewDefault(cid uint32) *Token {
	return New(cid, 0, defaultDuration)
}

// FromStr 从 token 字符串构造一个 *Token 变量
func FromStr(tokenStr string) (*Token, error) {
	token := &Token{}
	token.tokenStr = tokenStr

	sections := strings.Split(tokenStr, Delimiter)
	if len(sections) != tokenSectionAmount {
		return nil, ErrInvalidTokenStr
	}

	cid, err := strconv.Atoi(sections[0])
	if err != nil {
		return nil, ErrInvalidCID
	}
	token.CID = uint32(cid)

	ctlNum, err := strconv.Atoi(sections[1])
	if err != nil {
		return nil, ErrInvalidControl
	}
	token.Ctl = ControlNumber(uint32(ctlNum))

	expired, err := strconv.Atoi(sections[2])
	if err != nil {
		return nil, ErrInvalidExpired
	}
	token.Expired = uint32(expired)
	return token, nil
}

// Str 输出 token 字符串
func (t *Token) Str(key []byte) (string, error) {
	buf := bytes.NewBuffer(make([]byte, 0, 12))
	ctlNum := t.Ctl.Number()
	err := binary.Write(buf, binary.LittleEndian, t.CID)
	err = binary.Write(buf, binary.LittleEndian, ctlNum)
	err = binary.Write(buf, binary.LittleEndian, t.Expired)
	if err != nil {
		log.Printf("lytoken: write token buf failed: %v\n", err)
		return "", err
	}
	digestStr, err := hmacMD5Str(key, buf.Bytes())
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d_%d_%d_%s", t.CID, ctlNum, t.Expired, digestStr), nil
}

// OptionValue 获取指定控制选项的值
func (t *Token) OptionValue(name string) string {
	return t.Ctl.OptionValue(name)
}

// IsValid 返回 token 是否有效, 需要未过期且能通过校验
func (t *Token) IsValid(keys [][]byte) bool {
	if t.CID == 0 {
		return false
	}
	if t.IsExpired() {
		return false
	}
	for _, key := range keys {
		tokenStr, err := t.Str(key)
		if err != nil {
			return false
		}
		if t.tokenStr == tokenStr {
			return true
		}
	}
	return false

}

// IsExpired 检查 token 是否已经过期
func (t *Token) IsExpired() bool {
	return t.Expired < uint32(time.Now().Unix())
}

// ValidDuration  token 剩下的有效期, 单位秒, 可能会有误差
func (t *Token) ValidDuration() int64 {
	return int64(t.Expired) - int64(time.Now().Unix())
}

// hmacMD5Str hmac MD5 算法哈希 token 内容
func hmacMD5Str(key, content []byte) (string, error) {
	h := hmac.New(md5.New, key)
	_, err := h.Write(content)
	if err != nil {
		log.Printf("lytoken: hmac md5 failed: %v\n", err)
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func init() {
	initControlOptions()
}
