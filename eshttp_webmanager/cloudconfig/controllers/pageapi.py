#!/usr/bin/env python
# -*- coding:utf-8 -*-

import time
from django.shortcuts import render
from django.utils.dateformat import format
from django.http import HttpResponse
from cloudconfig.models import Template
from cloudconfig.models import Node
from cloudconfig.models import Status
from cloudconfig.helper import get_client_ip
from cloudconfig.helper import JsonResponse
from cloudconfig.helper import json_decode



def node_list(r):
    ret = []
    nodes = Node.objects.all()
    for node in nodes:
        ret.append({
            "id": node.id,
            "name": node.name,
            "ip": node.ip,
            "lastping": node.lastping,
            "confirmed": node.confirmed,
            "config_name": node.currentconfig.name if node.currentconfig else "Empty",
            "config_id": node.currentconfig.id if node.currentconfig else None,
            "lastchange": node.lastchange,
        })
    return JsonResponse(ret)

def node_edit(r):
    nodeid = r.JSON.get("id", None)
    name = r.JSON.get("name", None)
    config_id = r.JSON.get("config_id", None)

    if nodeid == None:
        return JsonResponse(None, False, "No node ID specificed")

    try:
        node = Node.objects.get(pk=nodeid)
        if name != None:
            node.name = name
        if config_id != None:
            node.currentconfig_id = int(config_id)
        node.save()
        return JsonResponse({"id": int(nodeid)})
    except Exception, e:
        return JsonResponse(None, False, str(e))

def node_confirm(r):
    nodeid = r.JSON.get("id", None)

    if nodeid == None:
        return JsonResponse(None, False, "No node ID specificed")

    try:
        node = Node.objects.get(pk=nodeid)
        node.confirmed = True
        node.save()
        return JsonResponse({"id": int(nodeid)})
    except Exception, e:
        return JsonResponse(None, False, str(e))

def node_delete(r):
    nodeid = r.JSON.get("id", None)

    if nodeid == None:
        return JsonResponse(None, False, "No node ID specificed")

    try:
        Node.objects.get(pk=nodeid).delete()
        return JsonResponse({"id": int(nodeid)})
    except Exception, e:
        return JsonResponse(None, False, str(e))









def config_list(r):
    ret = []
    templates = Template.objects.all()
    for template in templates:
        ret.append({
            "id": template.id,
            "name": template.name,
            "content": template.content,
            "created": format(template.created, "U"),
            "updated": format(template.updated, "U"),
        })
    return JsonResponse(ret)


def config_add(r):

    name = r.JSON.get("name", None)
    content = r.JSON.get("content", None)

    if name == None or content == None:
        return JsonResponse(None, False, "name or content invalid")

    try:
        template = Template(
            name=name,
            content=content,
        )
        template.save()
        return JsonResponse({"id": int(template.id)})
    except Exception, e:
        return JsonResponse(None, False, str(e))


def config_edit(r):
    templateid = r.JSON.get("id", None)
    name = r.JSON.get("name", None)
    content = r.JSON.get("content", None)

    if templateid == None:
        return JsonResponse(None, False, "No config ID specificed")

    try:
        template = Template.objects.get(pk=templateid)
        if name != None:
            template.name = name
        if content != None:
            template.content = content
        template.save()
        return JsonResponse({"id": int(templateid)})
    except Exception, e:
        return JsonResponse(None, False, str(e))

def config_delete(r):
    templateid = r.JSON.get("id", None)

    if templateid == None:
        return JsonResponse(None, False, "No config ID specificed")

    try:
        Template.objects.get(pk=templateid).delete()
        return JsonResponse({"id": int(templateid)})
    except Exception, e:
        return JsonResponse(None, False, str(e))





def status_list(r):
    ret = []
    nodes = Node.objects.filter(confirmed=True).all()
    for node in nodes:

        status = {}
        if node.status is not None:
            status["index_speed"] = {}
            status["index_speed"]["es_index_bytes_per_second"] = node.status.esindexspeed
            status["index_speed"]["s3_index_bytes_per_second"] = node.status.s3indexspeed
            status["system"] = {}
            status["system"]["last_reload_at"] = node.status.lastreload
            status["system"]["uptime"] = node.status.uptime
            status["system"]["updated_time"] = int(format(node.status.updated, "U"))
            status["system"]["updated_delay"] = time.time()-int(format(node.status.updated, "U"))
            status["http"] = {}
            status["http"]["cache"] = {}
            status["http"]["error"] = {}
            status["http"]["cache"]["es_cache_total"] = node.status.httpcacheestotal
            status["http"]["cache"]["es_cache_used"] = node.status.httpcacheesused
            status["http"]["cache"]["s3_cache_total"] = node.status.httpcaches3total
            status["http"]["cache"]["s3_cache_used"] = node.status.httpcaches3used
            status["http"]["error"]["counter_accepted"] = node.status.httpcounteraccepted
            status["http"]["error"]["counter_bad_parameter"] = node.status.httpcounterbadparam
            status["http"]["error"]["counter_invalid_json"] = node.status.httpcounterinvalidjson
            status["http"]["qps"] = node.status.httpcounterqps
            status["local_file_buffer"] = {}
            status["local_file_buffer"]["es_bytes"] = node.status.eslocalcache
            status["local_file_buffer"]["s3_bytes"] = node.status.s3localcache
            status["sender"] = {}
            status["sender"]["s3"] = json_decode(node.status.s3uploadstatus)
            status["sender"]["es"] = json_decode(node.status.esuploadstatus)

        ret.append({
            "id": node.id,
            "name": node.name,
            "ip": node.ip,
            "heartbeat": node.lastping,
            "status": status,
        })
    return JsonResponse(ret)
