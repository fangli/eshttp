#!/usr/bin/env python
# -*- coding:utf-8 -*-

from django.conf import settings
from django.http import HttpResponse
from django.utils import simplejson
import time


def get_client_ip(request):
    x_forwarded_for = request.META.get('HTTP_X_FORWARDED_FOR')
    if x_forwarded_for:
        ip = x_forwarded_for.split(',')[0]
    else:
        ip = request.META.get('REMOTE_ADDR')
    return ip



class JsonResponse(HttpResponse):
    """
        JSON response
    """
    def __init__(self, content, is_ok=True, message="", mimetype='application/json', status=None, content_type=None):
        time.sleep(0)
        content = {
            "result": "ok" if is_ok else "error",
            "message": message,
            "body": content,
        }
        super(JsonResponse, self).__init__(
            content=simplejson.dumps(content, ensure_ascii = False),
            mimetype=mimetype,
            status=status,
            content_type=content_type,
        )
        self['Access-Control-Allow-Origin'] = '*'

def json_decode(str):
    return simplejson.loads(str)
