package stream

import (
	"dev/monitor_oss_interface/common"
	"fmt"
	"github.com/lal/pkg/httpflv"
	"github.com/lal/pkg/rtmp"
	"go.uber.org/zap"
	"time"
)

func PushStream(flvFileName, rtmpPushURL string, timeout time.Duration, logger *zap.Logger) (err error) {
	var ffr httpflv.FlvFileReader
	err = ffr.Open(flvFileName)
	if err != nil {
		logger.Error(fmt.Sprintf("videoUpload [PushStream] open flv file %s failed, error %v.", flvFileName, err))
		return
	}
	defer ffr.Dispose()
	logger.Info(fmt.Sprintf("videoUpload [PushStream] open flv file %s success.", flvFileName))

	flvHeader, err := ffr.ReadFlvHeader()
	if err != nil {
		logger.Error(fmt.Sprintf("videoUpload [PushStream] read flv file %s failed, error %v.", flvFileName, err))
		return
	}
	logger.Info(fmt.Sprintf("videoUpload [PushStream] read flv header success, error %v.", flvHeader))
	ps := rtmp.NewPushSession(int64(timeout))
	err = ps.Push(rtmpPushURL)
	if err != nil {
		logger.Error(fmt.Sprintf("videoUpload [PushStream] push flv file %s failed, error %v.", flvFileName, err))
		return
	}
	logger.Info(fmt.Sprintf("videoUpload [PushStream]  start push flv file %s.", flvFileName))

	var prevTS uint32
	firstA := true
	firstV := true
	retryCount := 0
	currentTime := time.Now().Unix()
	flag := true
	for (retryCount < 5) {
		if ((time.Now().Unix() - 60) > currentTime) {
			break
		}
		tag, err := ffr.ReadTag()
		if err != nil {
			err = nil
			logger.Warn(fmt.Sprintf("videoUpload [PushStream] stream tag is empty, flv file %s.", flvFileName))
			break
		}

		// TODO chef: 转换代码放入lal某个包中
		var h rtmp.Header
		h.MsgLen = int(tag.Header.DataSize) //len(tag.Raw)-httpflv.TagHeaderSize
		h.Timestamp = int(tag.Header.Timestamp)
		h.MsgTypeID = int(tag.Header.T)
		h.MsgStreamID = rtmp.MSID1
		switch tag.Header.T {
		case httpflv.TagTypeMetadata:
			h.CSID = rtmp.CSIDAMF
		case httpflv.TagTypeAudio:
			h.CSID = rtmp.CSIDAudio
		case httpflv.TagTypeVideo:
			h.CSID = rtmp.CSIDVideo
		}

		// 把第一个音频和视频的时间戳改成0
		if tag.Header.T == httpflv.TagTypeAudio && firstA {
			h.Timestamp = 0
			firstA = false
		}
		if tag.Header.T == httpflv.TagTypeVideo && firstV {
			h.Timestamp = 0
			firstV = false
		}

		chunks := rtmp.Message2Chunks(tag.Raw[11:11+h.MsgLen], &h, rtmp.LocalChunkSize)

		// 第一个包直接发送
		if prevTS == 0 {
			err = ps.TmpWrite(chunks)
			if err != nil {
				flag = false
				logger.Error(fmt.Sprintf("videoUpload [PushStream] push rtmp first packet failed, error %v.", err))
				retryCount++
				time.Sleep(time.Second * 1)
				continue
			}
			prevTS = tag.Header.Timestamp
			continue
		}

		// 相等或回退了直接发送
		if tag.Header.Timestamp <= prevTS {
			err = ps.TmpWrite(chunks)
			if err != nil {
				flag = false
				logger.Error(fmt.Sprintf("videoUpload [PushStream] push rtmp packet failed, error %v.", err))
				retryCount++
				time.Sleep(time.Second * 1)
				continue
			}
			prevTS = tag.Header.Timestamp
			continue
		}

		if tag.Header.Timestamp > prevTS {
			diff := tag.Header.Timestamp - prevTS

			// 跳跃超过了30秒，直接发送
			if diff > 30000 {
				err = ps.TmpWrite(chunks)
				if err != nil {
					flag = false
					logger.Error(fmt.Sprintf("videoUpload [PushStream]over 30 seconds time out push rtmp packet failed, error %v.", err))
					retryCount++
					time.Sleep(time.Second * 1)
					continue
				}
				prevTS = tag.Header.Timestamp
				continue
			}

			// 睡眠后发送，睡眠时长为时间戳间隔
			time.Sleep(time.Duration(diff) * time.Millisecond)
			err = ps.TmpWrite(chunks)
			if err != nil {
				flag = false
				logger.Error(fmt.Sprintf("videoUpload [PushStream] after sleep some seconds push rtmp packet failed,error %v.", err))
				retryCount++
				time.Sleep(time.Second * 1)
				continue
			}
			prevTS = tag.Header.Timestamp
			continue
		}
	}
	if !flag && err == nil{
		err = common.CustomError{}
	}
	return
}
