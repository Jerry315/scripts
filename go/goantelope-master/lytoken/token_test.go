package lytoken

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	key1             = []byte("27cacd2597455c620c94b71aa1e12466")
	key2             = []byte("13c8b80ae0f524c3f85f856cc5c0dc83")
	keys             = [][]byte{key1, key2}
	cids             = []uint32{1, 10003775, 10000415}
	invalidTokenStrs = []struct {
		err      error
		tokenStr string
	}{
		{ErrInvalidTokenStr, "1_1_1_1_1_1_1_1"},
		{ErrInvalidCID, "abc_1_1_xxx"},
		{ErrInvalidControl, "1_abc_1_xxx"},
		{ErrInvalidExpired, "1_1_abc_xxx"},
	}

	tokenData = []struct {
		correct        bool
		alreadyExpired bool
		cid            uint32
		ctlNum         uint32
		duration       time.Duration
		expired        uint32
		key            []byte
		tokenStr       string
		optionValues   map[string]string
	}{
		// 过期但正确的 token 数据
		{
			correct:  true,
			cid:      1,
			ctlNum:   0,
			expired:  1529057636,
			key:      key1,
			tokenStr: "1_0_1529057636_9cb554edc06eefc4b311d86143e07f95",
			optionValues: map[string]string{
				CtlRTMPLive: "0",
				CtlHLSLive:  "0",
				CtlStorage:  "0000",
				CtlSound:    "0",
			},
		},
		{
			correct:  true,
			cid:      538443778,
			ctlNum:   65280,
			expired:  1529057636,
			key:      key2,
			tokenStr: "538443778_65280_1529057636_5d9ec7bb72c2caf501611e8313820276",
			optionValues: map[string]string{
				CtlRTMPLive:       "0",
				CtlHLSLive:        "0",
				CtlStorage:        "0000",
				CtlWatchPublic:    "1",
				CtlWatchPrivate:   "1",
				CtlWatchTimeShift: "1",
				CtlSoundPassBack:  "1",
				CtlVideoPassBack:  "1",
				CtlGetSnapshot:    "1",
				CtlSound:          "1",
			},
		},
		{
			correct:  true,
			cid:      1000315,
			ctlNum:   3223060480,
			expired:  1529057636,
			key:      key1,
			tokenStr: "1000315_3223060480_1529057636_94d0ec44de9503becd50f7dc00d3c8d3",
			optionValues: map[string]string{
				CtlRTMPLive:         "1",
				CtlHLSLive:          "1",
				CtlValidateIP:       "0",
				CtlValidateRefer:    "0",
				CtlStorage:          "0001",
				CtlUseCircleStorage: "1",
				CtlUseEventStorage:  "1",
				CtlSound:            "0",
			},
		},
		{
			correct:  true,
			cid:      2147550720,
			ctlNum:   4294967295,
			expired:  1529057636,
			key:      key2,
			tokenStr: "2147550720_4294967295_1529057636_cb4567c9ccbf602466af7b8981fd9164",
			optionValues: map[string]string{
				CtlRTMPLive: "1",
				CtlHLSLive:  "1",
				CtlStorage:  "1111",
				CtlSound:    "1",
			},
		},

		// 未过期的 token 数据，临时生成
		{
			correct:        true,
			alreadyExpired: false,
			cid:            10003775,
			ctlNum:         65280,
			duration:       24 * time.Hour,
			key:            key1,
		},
		{
			correct:        false,
			alreadyExpired: true,
			cid:            10003775,
			ctlNum:         65280,
			duration:       -24 * time.Hour,
			key:            key2,
		},
	}
)

func TestNewDefault(t *testing.T) {
	assert := assert.New(t)
	for _, cid := range cids {
		token := NewDefault(cid)
		assert.NotNil(token)
		assert.Equal(false, token.IsExpired())
		assert.Equal(uint32(0), token.Ctl.Number())
	}
}

func TestFromStr(t *testing.T) {
	assert := assert.New(t)
	for _, data := range tokenData {
		if !data.correct {
			continue
		}
		if data.tokenStr == "" {
			continue
		}
		if data.expired == 0 {
			continue
		}
		token, err := FromStr(data.tokenStr)

		assert.Nil(err)
		assert.NotNil(token)
		assert.Equal(data.cid, token.CID)
		assert.Equal(data.ctlNum, token.Ctl.Number())
		assert.Equal(data.expired, token.Expired)
	}

	for _, data := range invalidTokenStrs {
		token, err := FromStr(data.tokenStr)
		assert.Nil(token)
		assert.NotNil(err)
		assert.Equal(data.err, err)
	}
}

func TestStr(t *testing.T) {
	assert := assert.New(t)
	for _, data := range tokenData {
		if !data.correct {
			continue
		}
		if data.tokenStr == "" {
			continue
		}
		if data.expired == 0 {
			continue
		}
		token := &Token{
			CID:     data.cid,
			Ctl:     ControlNumber(data.ctlNum),
			Expired: data.expired,
		}
		t.Log("control number", token.Ctl.Number())
		tokenStr, err := token.Str(data.key)
		assert.Nil(err)
		assert.Equal(data.tokenStr, tokenStr)
	}
}

func TestOptionValue(t *testing.T) {
	assert := assert.New(t)
	for _, data := range tokenData {
		if len(data.optionValues) == 0 {
			continue
		}
		token := New(data.cid, data.ctlNum, data.duration)
		for option, value := range data.optionValues {
			v := token.OptionValue(option)
			assert.Equal(v, value)
		}
	}
}

func TestIsValid(t *testing.T) {
	assert := assert.New(t)
	for _, data := range tokenData {
		if data.tokenStr != "" {
			continue
		}
		token := New(data.cid, data.ctlNum, data.duration)
		tokenStr, err := token.Str(data.key)
		assert.Nil(err)
		assert.NotNil(token)

		token.tokenStr = tokenStr
		assert.Equal(data.correct, token.IsValid(keys))
	}

	invalidToken := New(0, 0, time.Second)
	tokenStr, err := invalidToken.Str(key1)
	assert.Nil(err)

	invalidToken.tokenStr = tokenStr
	assert.Equal(false, invalidToken.IsValid(keys))

	k1 := []byte("1")
	k2 := []byte("2")
	invalidKeys := [][]byte{k1, k2}
	invalidToken2 := New(1, 1, time.Hour)
	tokenStr, err = invalidToken2.Str(key1)
	assert.Nil(err)

	invalidToken2.tokenStr = tokenStr
	assert.Equal(false, invalidToken2.IsValid(invalidKeys))
}

func TestIsExpired(t *testing.T) {
	assert := assert.New(t)
	for _, data := range tokenData {
		if data.tokenStr != "" {
			continue
		}
		token := New(data.cid, data.ctlNum, data.duration)
		assert.NotNil(token)
		assert.Equal(data.alreadyExpired, token.IsExpired())
	}
}

func TestValidDuration(t *testing.T) {
	assert := assert.New(t)
	for _, data := range tokenData {
		if data.tokenStr != "" {
			continue
		}
		token := New(data.cid, data.ctlNum, data.duration)
		assert.NotNil(token)
		durationSecond := token.ValidDuration()
		t.Log("token expired", token.Expired)
		t.Log("token valid duration seconds", durationSecond)
		assert.Equal(data.alreadyExpired, durationSecond < 0)
	}
}
