#!/usr/bin/env python
# -*- coding: utf-8 -*-
import click
import json
import xlrd
from settings import FORMAT, ENDPOINT, METHOD, PROTOCOL, VERSION, CLIENT, RECORDS_FILE
from ddns_handle import request, AliRequest


def get_records(method, record_file):
    data = []
    readbook = xlrd.open_workbook(record_file)
    add = readbook.sheet_by_name(method)
    rows = add.nrows
    cols = add.ncols
    args = add.row_values(0, 0, cols)
    for r in range(1, rows):
        tmp = {}
        for c in range(cols):
            tmp[args[c]] = add.cell(r, c).value
        data.append(tmp)
    return data


def bytes_to_json(b):
    return json.dumps(json.loads(str(CLIENT.do_action_with_exception(b), encoding="utf-8")), indent=4)


@click.group()
def cli():
    pass


@click.command()
def get_all_domain():
    """
    获取当前账户下所有域名的信息
    """
    print(bytes_to_json(ar.get_all_domain()))


@click.command()
@click.argument("domain")
@click.argument("pg", default=1)
@click.argument("ps", default=20)
def get_one_domain(domain, pg, ps):
    """
    根据传入参数获取指定主域名的所有解析记录列表
    domain: 根域名
    pg: 如果返回页数较多，显示那一页
    ps: 每一页显示多少条记录
    """
    print(bytes_to_json(ar.get_one_domain_detail(domain, pg, ps)))


@click.command()
@click.argument("domain")
@click.argument("rr")
@click.argument("type_")
@click.argument("value")
def add_domain(domain, rr, type_, value):
    """
    添加一条解析记录
    domain: 根域名，baidu.com
    rr: 主机名，www
    type_: 解析类型，A记录、NS记录、MX记录等
    value: 记录值，A记录对应ip
    """
    print(bytes_to_json(ar.add_domain(domain, rr, type_, value)))


@click.command()
@click.argument("rdid")
@click.argument("rr")
@click.argument("type_")
@click.argument("value")
def update_domain(rdid, rr, type_, value):
    """
    更新一条解析记录
    rdid: 解析记录的ID，此参数在添加解析时会返回，在获取域名解析列表时会返回
    rr: 主机名，www
    type_: 解析类型，A记录、NS记录、MX记录等
    valaue: 记录值，A记录对应ip
    """
    print(bytes_to_json(ar.update_domain(rdid, rr, type_, value)))


@click.command()
@click.argument("rdid")
@click.argument("status")
def set_domain(rdid, status):
    """
    设置解析记录状态
    rdid: 解析记录的ID，此参数在添加解析时会返回，在获取域名解析列表时会返回
    status: Enable启用解析 Disable: 暂停解析
    """
    print(bytes_to_json(ar.set_domain(rdid, status)))


@click.command()
@click.argument("rdid")
def del_domain(rdid):
    """
    删除解析记录
    rdid: 解析记录的ID，此参数在添加解析时会返回，在获取域名解析列表时会返回
    """
    print(bytes_to_json(ar.del_domain(rdid)))


@click.command()
@click.argument("domain")
@click.argument("rr")
@click.argument("type_")
def del_sub_domain(domain, rr, type_):
    """
    删除主机记录对应的解析记录
    domain: 域名名称
    rr: 主机记录,www
    type_: 如果不填写，对应的全部解析记录类型。
    解析类型包括(不区分大小写)：A、MX、CNAME、TXT、REDIRECT_URL、FORWORD_URL、NS、AAAA、SRV
    """
    print(bytes_to_json(ar.del_sub_domain(domain, rr, type_)))


@click.command()
def batch_add_domain():
    """
    从模板文件中导入域名记录，批量添加
    """
    result_list = []
    records = get_records("add", RECORDS_FILE)
    for record in records:
        if record["Type"] == "A":
            record.pop("Priority")
        try:
            result = bytes_to_json(ar.do_action("AddDomainRecord", **record))
            record["RecordId"] = json.loads(result)["RecordId"]
            record["msg"] = "add record success"
            result_list.append(record)
        except Exception as e:
            print e
    print(json.dumps(result_list, indent=4))


@click.command()
def batch_update_domain():
    """
    从模板文件中导入域名记录，批量更新
    """
    result_list = []
    records = get_records("update", RECORDS_FILE)
    for record in records:
        result = bytes_to_json(ar.do_action("UpdateDomainRecord", **record))
        record["msg"] = "update record success"
        result_list.append(record)
    print(json.dumps(result_list, indent=4))


@click.command()
def batch_delete_domain():
    """
    从模板文件中导入域名记录，批量删除
    """
    result_list = []
    records = get_records("delete", RECORDS_FILE)
    for record in records:
        result = bytes_to_json(ar.do_action("DeleteDomainRecord", **record))
        record["msg"] = "delete record success"
        result_list.append(record)
    print(json.dumps(result_list, indent=4))


@click.command()
def batch_set_domain():
    """
    从模板文件中导入域名记录，批量禁用或启用
    """
    result_list = []
    records = get_records("set", RECORDS_FILE)
    for record in records:
        result = bytes_to_json(ar.do_action("SetDomainRecordStatus", **record))
        record["msg"] = "%s record success" % record["Status"]
        result_list.append(record)
    print(json.dumps(result_list, indent=4))


def main():
    cli.add_command(get_all_domain)
    cli.add_command(get_one_domain)
    cli.add_command(add_domain)
    cli.add_command(batch_add_domain)
    cli.add_command(update_domain)
    cli.add_command(batch_update_domain)
    cli.add_command(set_domain)
    cli.add_command(batch_set_domain)
    cli.add_command(del_domain)
    cli.add_command(batch_delete_domain)
    cli.add_command(del_sub_domain)
    cli()


if __name__ == '__main__':
    try:
        request = request(FORMAT, ENDPOINT, METHOD, PROTOCOL, VERSION, )
        ar = AliRequest(request)
        main()
    except Exception as e:
        print(e)
