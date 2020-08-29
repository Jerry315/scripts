package handle

import (
	"dev/check_storage_cycle/common"
	"fmt"
	"github.com/tealeg/xlsx"
	"go.uber.org/zap"
	"strconv"
)

func WriteExcel(name string,uncids []common.UnCid,logger *zap.Logger)  {
	exHandle := xlsx.NewFile()
	esHeaders := []string{
		"CID","设备名","分组","设备SN","厂商","型号","软件版本","build时间","通配视频周期","对象存储视频周期","通配图片周期","对象存储图片周期",
	}
	sheet,err := exHandle.AddSheet("cid")
	if err != nil{
		logger.Error(fmt.Sprintf("创建sheet失败. %v",err))
	}
	headerRow := sheet.AddRow()
	for _,header := range esHeaders{
		tmpCell := headerRow.AddCell()
		tmpCell.Value = header
	}
	for _,uncid := range uncids{
		tmpRow := sheet.AddRow()
		cidCell := tmpRow.AddCell()
		cid := strconv.Itoa(int(uncid.CID))
		cidCell.Value = cid
		nameCell := tmpRow.AddCell()
		nameCell.Value = uncid.Name
		groupCell := tmpRow.AddCell()
		groupCell.Value = uncid.Group
		snCell := tmpRow.AddCell()
		snCell.Value = uncid.SN
		brandCell := tmpRow.AddCell()
		brandCell.Value = uncid.Brand
		modelCell := tmpRow.AddCell()
		modelCell.Value = uncid.Model
		softCell := tmpRow.AddCell()
		softCell.Value = uncid.SoftwareVersion
		buildCell := tmpRow.AddCell()
		buildCell.Value = uncid.SoftwareBuild
		mvCell := tmpRow.AddCell()
		mVideo := strconv.Itoa(int(uncid.MVideo))
		mvCell.Value = mVideo
		ovCell := tmpRow.AddCell()
		oVideo := strconv.Itoa(int(uncid.OVideo))
		ovCell.Value = oVideo
		mpCell := tmpRow.AddCell()
		mp := strconv.Itoa(int(uncid.MPIC))
		mpCell.Value = mp
		opCell := tmpRow.AddCell()
		op := strconv.Itoa(int(uncid.OPIC))
		opCell.Value = op
	}
	err = exHandle.Save(name)
	if err != nil{
		logger.Error(fmt.Sprintf("create excel failed. %v",err))
	}else {
		logger.Info("create excel success.")
	}

}