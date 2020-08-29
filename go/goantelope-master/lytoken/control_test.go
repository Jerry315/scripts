package lytoken

import (
	"log"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	optionValues = map[string]bool{
		"110001": true,
		"120001": false,
		"000000": true,
		"111111": true,
		"12":     false,
		"0":      true,
		"1":      true,
		"abc":    false,
	}
	controlNumbersOptions = map[uint32][]CtlOption{
		uint32(0): []CtlOption{},
		uint32(65280): []CtlOption{
			OptionRTMPLive.Value("0"),
			OptionHLSLive.Value("0"),
			OptionWatchPublic.Value("1"),
			OptionWatchPrivate.Value("1"),
			OptionWatchTimeShift.Value("1"),
			OptionWatchVideo.Value("1"),
			OptionSoundPassBack.Value("1"),
			OptionVideoPassBack.Value("1"),
			OptionGetSnapshot.Value("1"),
			OptionSound.Value("1"),
		},
		uint32(3223060480): []CtlOption{
			OptionRTMPLive.Value("1"),
			OptionHLSLive.Value("1"),
			OptionValidateIP.Value("0"),
			OptionValidateRefer.Value("0"),
			OptionStorage.Value("0001"),
			OptionUseCircleStorage.Value("1"),
			OptionUseEventStorage.Value("1"),
			OptionAllowObjectDownload.Value("0"),
			OptionValidateObjectID.Value("0"),
		},
		uint32(4294967295): []CtlOption{
			OptionRTMPLive.Value("1"),
			OptionHLSLive.Value("1"),
			OptionValidateIP.Value("1"),
			OptionValidateRefer.Value("1"),
			OptionAcceptUDP.Value("1"),
			OptionUploadObject.Value("1"),
			OptionStorage.Value("1111"),
			OptionUseCircleStorage.Value("1"),
			OptionUseEventStorage.Value("1"),
			OptionAllowObjectDownload.Value("1"),
			OptionValidateObjectID.Value("1"),
			OptionWatchPublic.Value("1"),
			OptionWatchPrivate.Value("1"),
			OptionWatchTimeShift.Value("1"),
			OptionWatchVideo.Value("1"),
			OptionSoundPassBack.Value("1"),
			OptionVideoPassBack.Value("1"),
			OptionGetSnapshot.Value("1"),
			OptionSound.Value("1"),
			CtlOption{
				Name:     "len-2-reserved",
				position: 6,
				len:      2,
				value:    "11",
			},
			CtlOption{
				Name:     "len-4-reserved",
				position: 24,
				len:      8,
				value:    "11111111",
			},
		},
	}
	controls = map[uint32]Control{
		uint32(0): Control{
			bytes: []byte("00000000000000000000000000000000"),
		},
		uint32(65280): Control{
			bytes: []byte("00000000000000001111111100000000"),
		},
		uint32(3223060480): Control{
			bytes: []byte("11000000000111000000000000000000"),
		},
		uint32(4294967295): Control{
			bytes: []byte("11111111111111111111111111111111"),
		},
	}
	options = []struct {
		option   CtlOption
		newValue string
		valid    bool
	}{
		{CtlOption{"t1", 0, 1, "0"}, "1", true},
		{CtlOption{"t2", 2, 1, "0"}, "1", true},
		{CtlOption{"t3", 3, 9, "000000000"}, "111111111", true},
		{CtlOption{"t4", 23, 4, "1111"}, "0000", true},
		{CtlOption{"t5", 42, 4, "abcd"}, "efgh", false},
		{CtlOption{"t6", 47, 12, "1"}, "1", false},
	}
	testOptionList = []CtlOption{
		OptionRTMPLive, OptionHLSLive, OptionValidateIP, OptionValidateRefer,
		OptionAcceptUDP, OptionUploadObject, OptionStorage, OptionUseCircleStorage,
		OptionUseEventStorage, OptionAllowObjectDownload, OptionValidateObjectID,
		OptionWatchPublic, OptionWatchPrivate, OptionWatchTimeShift, OptionWatchVideo,
		OptionSoundPassBack, OptionVideoPassBack, OptionGetSnapshot, OptionSound,
	}
	testStorageOptions = []struct {
		expireType int
		value      string
		position   int
		ok         bool
	}{
		{1, "0001", 8, true},
		{2, "0010", 8, true},
		{4, "0100", 8, true},
		{8, "1000", 8, true},
		{16, "", 8, false},
	}
)

func TestNewControl(t *testing.T) {
	assert := assert.New(t)

	control := NewControl()
	assert.NotNil(control)
	assert.Len(control.bytes, 32)
	assert.Len(control.options, 0)
	assert.Equal(strings.Repeat("0", 32), string(control.bytes))
}

func TestControlOptionsOptionByName(t *testing.T) {
	assert := assert.New(t)

	for _, option := range testOptionList {
		opt, ok := Options.OptionByName(option.Name)
		assert.Equal(true, ok)
		assert.Equal(option.Name, opt.Name)
		assert.Equal(option.len, opt.len)
		assert.Equal(option.position, opt.position)
	}
}

func TestControlOptionsOptionByPosition(t *testing.T) {
	assert := assert.New(t)

	for _, option := range testOptionList {
		opt, ok := Options.OptionByPosition(option.position)
		assert.Equal(true, ok)
		assert.Equal(option.Name, opt.Name)
		assert.Equal(option.len, opt.len)
		assert.Equal(option.position, opt.position)
	}
}

func TestControlOptionsStorageOptionByExpireType(t *testing.T) {
	assert := assert.New(t)

	for _, data := range testStorageOptions {
		opt, ok := Options.StorageOptionByExpireType(data.expireType)
		assert.Equal(data.ok, ok)
		if !data.ok {
			continue
		}
		assert.Equal(data.value, opt.value)
		assert.Equal(data.position, opt.position)
	}
}

func TestControlNumber(t *testing.T) {
	assert := assert.New(t)
	for number, options := range controlNumbersOptions {
		control := ControlNumber(number)
		assert.NotNil(control)

		for _, option := range options {
			_option, ok := control.options[option.Name]
			if _, _ok := Options.OptionByName(option.Name); !_ok {
				continue
			}
			assert.Equal(true, ok)
			assert.Equal(option.position, _option.position)
			assert.Equal(option.value, _option.value)
		}
	}
}

func TestControl(t *testing.T) {
	assert := assert.New(t)
	for number, options := range controlNumbersOptions {
		control := NewControl()
		for _, option := range options {
			control = control.SetOption(option, option.value)
			assert.NotNil(control)
		}
		for _, option := range options {
			value := control.OptionValue(option.Name)
			assert.Equal(option.value, value)
		}
		log.Printf("control %v binary string %v\n", number, string(control.bytes))
		assert.Equal(number, control.Number())
	}
}

func TestControlGetNumber(t *testing.T) {
	assert := assert.New(t)

	for number, control := range controls {
		assert.Equal(number, control.Number())
	}

	control := NewControl()
	control.bytes[0] = byte('a')
	assert.Equal(uint32(0), control.Number())
}

func TestControlOptionValue(t *testing.T) {
	assert := assert.New(t)
	control := NewControl()
	assert.Equal("0", control.OptionValue(CtlRTMPLive))
	assert.Equal("0000", control.OptionValue(CtlStorage))
	assert.Equal("", control.OptionValue("notexist"))
}

func TestControlSetOption(t *testing.T) {
	assert := assert.New(t)
	control := NewControl()
	control.SetOption(OptionRTMPLive, "1")
	assert.Equal("1", control.OptionValue(OptionRTMPLive.Name))
}

func TestControlSetStorageByExpireType(t *testing.T) {
	assert := assert.New(t)
	control := NewControl()
	for _, data := range testStorageOptions {
		control.SetStorageByExpireType(data.expireType)
		if data.ok {
			assert.Equal(data.value, control.OptionValue(OptionStorage.Name))
		} else {
			assert.Equal("0000", control.OptionValue(OptionStorage.Name))
		}
	}
}

func TestCtlOptionSetValue(t *testing.T) {
	assert := assert.New(t)
	for _, optionData := range options {
		option := optionData.option.SetValue(optionData.newValue)
		if !optionData.valid {
			assert.Equal(optionData.option.value, option.value)
			continue
		}
		assert.Equal(optionData.newValue, option.value)
	}
}

func TestCtlOptionCopy(t *testing.T) {
	assert := assert.New(t)
	for _, optionData := range options {
		oldOpt := optionData.option
		newOpt := oldOpt.Copy()
		assert.Equal(oldOpt.Name, newOpt.Name)
		assert.Equal(oldOpt.len, newOpt.len)
		assert.Equal(oldOpt.position, newOpt.position)
		assert.Equal(oldOpt.value, newOpt.value)
	}
}

func TestCtlOptionValue(t *testing.T) {
	assert := assert.New(t)
	for _, optionData := range options {
		if !optionData.valid {
			continue
		}
		oldOpt := optionData.option
		newOpt := oldOpt.Value(optionData.newValue)
		assert.Equal(oldOpt.Name, newOpt.Name)
		assert.Equal(oldOpt.len, newOpt.len)
		assert.Equal(oldOpt.position, newOpt.position)
		assert.Equal(optionData.newValue, newOpt.value)
	}
}

func TestValidOptionValue(t *testing.T) {
	assert := assert.New(t)

	for value, rt := range optionValues {
		ok := validCtlOptionValue(value)
		assert.Equal(ok, rt)
	}
}
