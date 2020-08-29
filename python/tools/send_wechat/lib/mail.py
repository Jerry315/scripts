#!/usr/bin/env python
# -*- coding: utf-8 -*-
import smtplib
from email.mime.multipart import MIMEMultipart
from email.mime.text import MIMEText

from settings import smtp_server, smtp_user, smtp_passwd


class SendMail(object):
    def __init__(self):
        self._smtp_server = smtp_server
        self._mail_user = smtp_user
        self._mail_passwd = smtp_passwd
        self._type = 'plain'
        self.server = smtplib.SMTP_SSL(self._smtp_server, 465)
        self.server.login(self._mail_user, self._mail_passwd)

    def send_plain(self, to, subject, msg):
        message = MIMEText(msg, self._type, 'utf-8')
        message['Subject'] = subject
        message['From'] = self._mail_user
        message['TO'] = ';'.join(to)
        self.server.sendmail(self._mail_user, to, message.as_string())

    def send_file(self,to,subject,text,files):
        message = MIMEMultipart()
        message['Subject'] = subject
        message['From'] = self._mail_user
        message['TO'] = ';'.join(to)
        message.attach(MIMEText(text,'plain','utf-8'))
        att = MIMEText(open(files,'rb').read(),'base64','utf-8')
        att["Content-Type"] = 'application/octet-stream'
        att["Content-Disposition"] = 'attachment; filename="check_device_stats.txt"'
        message.attach(att)
        self.server.sendmail(self._mail_user,to,message.as_string())