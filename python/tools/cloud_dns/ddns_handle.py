#!/usr/bin/env python
# -*- coding: utf-8 -*-
from aliyunsdkcore.request import CommonRequest


def request(fmt, enpoint, method, protocol, version):
    req = CommonRequest()
    req.set_accept_format(fmt)
    req.set_domain(enpoint)
    req.set_method(method)
    req.set_protocol_type(protocol)
    req.set_version(version)
    return req


class AliRequest(object):
    def __init__(self, request):
        self.request = request

    def do_action(self, action, **kwargs):
        """
        一个公共接口，可以衍生后续所有的方法
        :param action: 对应操作方法
        :param kwargs: 需要传入的参数
        :return:
        """
        for key, value in kwargs.items():
            self.request.add_query_param(key, value)
        self.request.set_action_name(action)
        return self.request

    def get_all_domain(self):
        """
        获取当前账户下所有域名的信息
        :return:
        """
        self.request.set_action_name("DescribeDomains")
        return self.request

    def get_one_domain_detail(self, domain, pg=1, ps=20):
        """
        根据传入参数获取指定主域名的所有解析记录列表
        :param domain: 根域名
        :param pg: 如果返回页数较多，显示那一页
        :param ps: 每一页显示多少条记录
        :return:
        """
        self.request.add_query_param('DomainName', domain)
        self.request.add_query_param("PageNumber", pg)
        self.request.add_query_param("PageSize", ps)
        self.request.set_action_name("DescribeDomainRecords")
        return self.request

    def add_domain(self, domain, rr, type_, value):
        """
        添加一条解析记录
        :param domain: 根域名，baidu.com
        :param rr: 主机名，www
        :param type_: 解析类型，A记录、NS记录、MX记录等
        :param value: 	记录值，A记录对应ip
        :return:
        """
        self.request.add_query_param('DomainName', domain)
        self.request.add_query_param("RR", rr)
        self.request.add_query_param("Type", type_)
        self.request.add_query_param("Value", value)
        self.request.set_action_name("AddDomainRecord")
        return self.request

    def update_domain(self, RecordId, rr, type_, valaue):
        """
        更新一条解析记录
        :param RecordId: 解析记录的ID，此参数在添加解析时会返回，在获取域名解析列表时会返回
        :param rr: 主机名，www
        :param type_: 解析类型，A记录、NS记录、MX记录等
        :param valaue: 	记录值，A记录对应ip
        :return:
        """
        self.request.add_query_param('RecordId', RecordId)
        self.request.add_query_param("RR", rr)
        self.request.add_query_param("Type", type_)
        self.request.add_query_param("Value", valaue)
        self.request.set_action_name("UpdateDomainRecord")
        return self.request

    def set_domain(self, RecordId, status="Enable"):
        '''
        设置解析记录状态
        :param RecordId: 解析记录的ID，此参数在添加解析时会返回，在获取域名解析列表时会返回
        :param status: Enable: 启用解析 Disable: 暂停解析
        :return:
        '''
        self.request.add_query_param("RecordId", RecordId)
        self.request.add_query_param("Status", status)
        self.request.set_action_name("SetDomainRecordStatus")
        return self.request

    def del_domain(self, RecordId):
        '''
        删除解析记录
        :param RecordId: 解析记录的ID，此参数在添加解析时会返回，在获取域名解析列表时会返回
        :return:
        '''
        self.request.add_query_param("RecordId", RecordId)
        self.request.set_action_name("DeleteDomainRecord")
        return self.request

    def del_sub_domain(self, domain, rr, type_=None):
        """
        删除主机记录对应的解析记录,根据传入参数删除主机记录对应的解析记录。
        如果被删除的解析记录中存在锁定解析，则该锁定解析不会被删除。
        :param domain: 域名名称
        :param rr: 主机记录,www
        :param type_: 如果不填写，则返回子域名对应的全部解析记录类型。
        解析类型包括(不区分大小写)：A、MX、CNAME、TXT、REDIRECT_URL、FORWORD_URL、NS、AAAA、SRV
        :return:
        """
        self.request.add_query_param("DomainName", domain)
        self.request.add_query_param("RR", rr)
        self.request.add_query_param("Type", type_)
        self.request.set_action_name("DeleteSubDomainRecords")
        return self.request
