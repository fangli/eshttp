#!/usr/bin/env python
# -*- coding:utf-8 -*-

import time
from django.utils import simplejson
from django.shortcuts import render
from django.http import HttpResponse
from cloudconfig.models import Template
from cloudconfig.models import Node
from cloudconfig.models import Status
from cloudconfig.helper import get_client_ip
from cloudconfig.helper import JsonResponse

def getConfig(r):
    ip = get_client_ip(r)
    cli = Node.objects.filter(ip=ip)

    if cli.count() == 0:
        new_node = Node(
            name="N/A",
            ip=ip,
            lastping=time.time(),
            confirmed=False,
            currentconfig=None,
            lastchange=time.time(),
        )
        new_node.save()
        return HttpResponse("NO CONFIG ERROR")
    else:
        cli = cli[0]
        cli.lastping=time.time()
        cli.save()
        if cli.confirmed == True:
            if cli.currentconfig != None:
                return HttpResponse(cli.currentconfig.content)
            else:
                return HttpResponse("EMPTY CONFIG")
        else:
            return HttpResponse("NOT CONFIRM")

def postStatus(r):
    ip = get_client_ip(r)
    cli = Node.objects.filter(ip=ip)

    if cli.count() == 0:
        return HttpResponse("NOT FOUND")

    cli = cli[0]
    if cli.status == None:
        status = Status(
            esindexspeed=r.JSON.get("es_indexer", {}).get("bytes_per_second", -1),
            s3indexspeed=r.JSON.get("s3_indexer", {}).get("bytes_per_second", -1),
            lastreload=r.JSON.get("system", {}).get("last_reload_at", -1),
            uptime=r.JSON.get("system", {}).get("uptime", -1),
            httpcacheestotal=r.JSON.get("http", {}).get("cache", {}).get("es_total", -1),
            httpcacheesused=r.JSON.get("http", {}).get("cache", {}).get("es_used", -1),
            httpcaches3total=r.JSON.get("http", {}).get("cache", {}).get("s3_total", -1),
            httpcaches3used=r.JSON.get("http", {}).get("cache", {}).get("s3_used", -1),
            httpcounteraccepted=r.JSON.get("http", {}).get("counter", {}).get("accepted", -1),
            httpcounterbadparam=r.JSON.get("http", {}).get("counter", {}).get("bad_params", -1),
            httpcounterinvalidjson=r.JSON.get("http", {}).get("counter", {}).get("invalid_json", -1),
            httpcounterqps=r.JSON.get("http", {}).get("counter", {}).get("qps", -1),
            eslocalcache=r.JSON.get("es_uploader", {}).get("file_buffer_size", -1),
            s3localcache=r.JSON.get("s3_uploader", {}).get("file_buffer_size", -1),
            s3uploadstatus=simplejson.dumps(r.JSON.get("s3_uploader", {}).get("upload_status", []), ensure_ascii = False),
            esuploadstatus=simplejson.dumps(r.JSON.get("es_uploader", {}).get("upload_status", []), ensure_ascii = False),
        )
        status.save()
        cli.status = status
        cli.save()
    else:
        status = Status.objects.get(pk=cli.status.id)
        status.esindexspeed=r.JSON.get("es_indexer", {}).get("bytes_per_second", -1)
        status.s3indexspeed=r.JSON.get("s3_indexer", {}).get("bytes_per_second", -1)
        status.lastreload=r.JSON.get("system", {}).get("last_reload_at", -1)
        status.uptime=r.JSON.get("system", {}).get("uptime", -1)
        status.httpcacheestotal=r.JSON.get("http", {}).get("cache", {}).get("es_total", -1)
        status.httpcacheesused=r.JSON.get("http", {}).get("cache", {}).get("es_used", -1)
        status.httpcaches3total=r.JSON.get("http", {}).get("cache", {}).get("s3_total", -1)
        status.httpcaches3used=r.JSON.get("http", {}).get("cache", {}).get("s3_used", -1)
        status.httpcounteraccepted=r.JSON.get("http", {}).get("counter", {}).get("accepted", -1)
        status.httpcounterbadparam=r.JSON.get("http", {}).get("counter", {}).get("bad_params", -1)
        status.httpcounterinvalidjson=r.JSON.get("http", {}).get("counter", {}).get("invalid_json", -1)
        status.httpcounterqps=r.JSON.get("http", {}).get("counter", {}).get("qps", -1)
        status.eslocalcache=r.JSON.get("es_uploader", {}).get("file_buffer_size", -1)
        status.s3localcache=r.JSON.get("s3_uploader", {}).get("file_buffer_size", -1)
        status.s3uploadstatus=simplejson.dumps(r.JSON.get("s3_uploader", {}).get("upload_status", []), ensure_ascii = False)
        status.esuploadstatus=simplejson.dumps(r.JSON.get("es_uploader", {}).get("upload_status", []), ensure_ascii = False)
        status.save()

    return HttpResponse("OK")
