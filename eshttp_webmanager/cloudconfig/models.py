#!/usr/bin/env python
# -*- coding:utf-8 -*-

from django.db import models

class Template(models.Model):
    name = models.CharField(max_length=200, unique=True)
    content = models.TextField()
    created = models.DateTimeField(auto_now_add=True)
    updated = models.DateTimeField(auto_now=True)

class Status(models.Model):
    esindexspeed = models.IntegerField(max_length=50)
    s3indexspeed = models.IntegerField(max_length=50)
    lastreload = models.IntegerField(max_length=50)
    uptime = models.IntegerField(max_length=50)
    httpcacheestotal = models.IntegerField(max_length=50)
    httpcacheesused = models.IntegerField(max_length=50)
    httpcaches3total = models.IntegerField(max_length=50)
    httpcaches3used = models.IntegerField(max_length=50)
    httpcounteraccepted = models.IntegerField(max_length=50)
    httpcounterbadparam = models.IntegerField(max_length=50)
    httpcounterinvalidjson = models.IntegerField(max_length=50)
    httpcounterqps = models.IntegerField(max_length=10)
    eslocalcache = models.IntegerField(max_length=50)
    s3localcache = models.IntegerField(max_length=50)
    s3uploadstatus = models.TextField()
    esuploadstatus = models.TextField()
    updated = models.DateTimeField(auto_now=True)

class Node(models.Model):
    name = models.CharField(max_length=200)
    ip = models.CharField(max_length=15, unique=True)
    lastping = models.IntegerField(max_length=10)
    confirmed = models.BooleanField(max_length=200, default=False)
    currentconfig = models.ForeignKey(Template, null=True, on_delete=models.PROTECT)
    status = models.ForeignKey(Status, null=True, on_delete=models.PROTECT)
    lastchange = models.IntegerField(max_length=10)
