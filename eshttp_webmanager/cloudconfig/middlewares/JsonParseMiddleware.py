#!/usr/bin/env python
# -*- coding:utf-8 -*-

from django.utils import simplejson
from django import http


class JsonParseMiddleware(object):
    """
    This middleware parse POST JSON body if application/json is offered in the
    header
    """

    def __init__(self):
        pass

    def _is_ajax(self, request):
        if 'CONTENT_TYPE' in request.META and "json" in request.META['CONTENT_TYPE']:
            return True
        else:
            return False

    def process_request(self, request):
        if request.META['REQUEST_METHOD'] == "POST":
            if self._is_ajax(request):
                try:
                    request.JSON = simplejson.loads(request.body)
                except:
                    return http.HttpResponseBadRequest('<h1>Unable to decode the JSON request body</h1>')
            else:
                request.JSON = request.POST
        else:
            request.JSON = request.GET
        return None
