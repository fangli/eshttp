#!/usr/bin/env python
# -*- coding:utf-8 -*-

from django.conf.urls import patterns, url

from cloudconfig.controllers import page
from cloudconfig.controllers import pageapi
from cloudconfig.controllers import api

urlpatterns = patterns('',

    # Pages
    url(r'^/?$', page.index),

    # API for pages
    url(r'^api/node/?$', pageapi.node_list),
    url(r'^api/node/edit?$', pageapi.node_edit),
    url(r'^api/node/confirm?$', pageapi.node_confirm),
    url(r'^api/node/delete?$', pageapi.node_delete),

    url(r'^api/config/?$', pageapi.config_list),
    url(r'^api/config/edit?$', pageapi.config_edit),
    url(r'^api/config/add?$', pageapi.config_add),
    url(r'^api/config/delete?$', pageapi.config_delete),

    url(r'^api/status/?$', pageapi.status_list),

    # API for nodes
    url(r'^node/config/?$', api.getConfig),
    url(r'^node/status/?$', api.postStatus),
)
