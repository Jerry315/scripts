# -*- coding: utf-8 -*-
import xlwt
import time
import threading
from datetime import datetime, timedelta
from settings import report_file, config
from common import Logger

logger = Logger()


def create_assist_date(start_time=None, end_time=None):
    # 创建日期列表

    datestart = datetime.fromtimestamp(start_time).strftime("%Y-%m-%d")
    dateend = datetime.fromtimestamp(end_time).strftime("%Y-%m-%d")

    # 转换为日期格式
    datestart = datetime.strptime(datestart, "%Y-%m-%d")
    dateend = datetime.strptime(dateend, "%Y-%m-%d")
    date_list = []
    while datestart <= dateend:
        tmp_date = datestart.strftime("%Y-%m-%d")
        date_list.append(tmp_date)
        datestart += timedelta(days=1)
    return date_list


def set_style(name, height, bold=False, left=False):
    style = xlwt.XFStyle()  # 初始化样式

    # 设置字体
    font = xlwt.Font()  # 为样式创建字体
    font.name = name  # 'Calibri'
    font.bold = bold
    font.color_index = 4
    font.height = height
    style.font = font

    # 设置居中
    alignment = xlwt.Alignment()
    if left:
        alignment.horz = xlwt.Alignment.HORZ_LEFT  # 水平居中
    else:
        alignment.horz = xlwt.Alignment.HORZ_CENTER
    alignment.vert = xlwt.Alignment.VERT_CENTER  # 垂直居中
    style.alignment = alignment

    return style


def write_excel(func, static_data, start_time, end_time, *args, **kwargs):
    f = xlwt.Workbook(encoding='utf-8')  # 创建工作簿
    static_field = ["摄像机名字", "安装分组", "CID", "S/N", "是否绑定"]
    if "is_bind" in kwargs and kwargs["is_bind"]:
        new_static_data = {"data": []}
        for item in static_data["data"]:
            if item.get("is_bind", False):
                new_static_data["data"].append(item)
        static_data = new_static_data
    # 生成周期表头
    date_list = create_assist_date(start_time, end_time)
    period_field = ["%s在线率" % d for d in date_list]

    row0 = u"南昌云摄像机在线率统计周报\n%s~%s" % (date_list[0], date_list[-1])

    # 创建在线率周报
    sheet2 = f.add_sheet(u"在线率周报", cell_overwrite_ok=True)
    sheet2_row1 = """备注：\n\t1、数据取自云平台的自动记录，原始记录时间点为每台摄像机每天0时~24时，
               每小时一次，共计24个采样点。摄像机处于“推流中”（即：有视频云存储、且随时可开流查看）状态则视为在线。\n\t2、只统计
               了能在小羚通app上能看到的摄像机设备。（不包含单兵、机器人、CIG接入、国标接入的摄像机）\n\t3、表中每日在线率为当天记录
               的平均值，当天在线率计算方法：(∑(每日有效统计的在线次数) / ∑(每日有效统计的总采样数))，已排除统计周期内数据不全的
               影响。例如，一天24条在线状态中，有20个是在线，则平均在线率为20/24*100%=83.3%"""

    # 创建表头
    sheet2.write_merge(0, 0, 0, len(static_field) + len(period_field) - 1, row0,
                       set_style('Calibri', 220, True))
    sheet2.write_merge(1, 1, 0, len(static_field) + len(period_field) - 1, sheet2_row1,
                       set_style('Calibri', 220, False, True))

    for i in range(len(static_field + period_field)):
        sheet2.write(2, i, (static_field + period_field)[i], set_style('Calibri', 220))
        if len((static_field + period_field)[i]) <= 15:
            sheet2.col(i).width = 256 * 15
        else:
            sheet2.col(i).width = 256 * len((static_field + period_field)[i])

    # 创建原始数据
    sheet1 = f.add_sheet(u"原始数据", cell_overwrite_ok=True)

    # 创建表头
    sheet1.write_merge(0, 0, 0, len(static_field) - 1, row0, set_style('Calibri', 220, True))

    # 创建第二行
    for i in range(len(static_field)):
        sheet1.write_merge(1, 2, i, i, static_field[i], set_style('Calibri', 220))
        if len(static_field[i]) <= 15:
            sheet1.col(i).width = 256 * 15
        else:
            sheet1.col(i).width = 256 * len(static_field[i])

    # 生成合并表头
    c, n = 5, 0
    while c <= len(period_field) * 24 + 4 and n < len(period_field):
        sheet1.write_merge(1, 1, c, c + 24 - 1, period_field[n], set_style('Calibri', 220))

        for j in range(24):
            sheet1.write(2, c + j, j, set_style('Calibri', 220))
            sheet1.col(c + j).width = 256 * 5
        c += 24
        n += 1

    # 插入数据，每次插入固定长度的数据
    s = 0
    threads = []
    step = config['step']
    while (len(static_data["data"]) - s) > 0:
        if (len(static_data["data"]) - s) < step:
            step = len(static_data["data"]) - s
        t = threading.Thread(
            target=insert_data(func, sheet1, sheet2, static_data["data"][s:s + step], start_time, end_time, date_list,
                               s))
        threads.append(t)
        s += step

    logger.info("insert data thread starting")
    for t in threads:
        t.start()

    logger.info("insert data thread working")
    for t in threads:
        t.join()

    logger.info("insert data thread finish")
    f.save(report_file)


def insert_data(func, sheet1, sheet2, item, start_time, end_time, date_list, s):
    cid_list = [value["cid"] for value in item]
    for i in range(len(item)):
        sheet1.write(3 + i + s, 0, item[i].get("name", ""))
        sheet1.write(3 + i + s, 1, item[i].get("group", ""))
        sheet1.write(3 + i + s, 2, item[i].get("cid", ""))
        sheet1.write(3 + i + s, 3, item[i].get("sn", ""))
        sheet1.write(3 + i + s, 4, item[i].get("is_bind", ""))
        sheet2.write(3 + i + s, 0, item[i].get("name", ""))
        sheet2.write(3 + i + s, 1, item[i].get("group", ""))
        sheet2.write(3 + i + s, 2, item[i].get("cid", ""))
        sheet2.write(3 + i + s, 3, item[i].get("sn", ""))
        sheet2.write(3 + i + s, 4, item[i].get("is_bind", ""))

    period_data = func(cid_list, start_time, end_time)

    ss = 5
    for p in range(len(date_list)):
        zero_time = time.mktime(time.strptime(date_list[p], '%Y-%m-%d'))
        last_time = zero_time + 86400 - 1
        percent_list = [[0, 0] for pp in range(len(cid_list))]
        for pd in period_data:
            if zero_time <= pd['create_time'] <= last_time:
                sheet1.write(3 + s + cid_list.index(pd['cid']), ss + pd['tm_hour'], pd['status'])
                if pd['status'] == "Y":
                    percent_list[cid_list.index(pd['cid'])][0] += 1
                else:
                    percent_list[cid_list.index(pd['cid'])][1] += 1

        for t in range(len(percent_list)):
            if not (percent_list[t][0] + percent_list[t][1]):
                percent = '0.00%'
            else:
                percent = '%.2f%%' % (float(percent_list[t][0]) / float(percent_list[t][0] + percent_list[t][1]) * 100)

            sheet2.write(3 + s + t, 5 + p, percent)
        ss += 24
