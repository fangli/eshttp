#!/usr/bin/env python
# -*- coding:utf-8 -*-

from django.conf.urls import patterns, include, url

urlpatterns = patterns('',
    # Examples:
    # url(r'^$', 'eshttp_webmanager.views.home', name='home'),
    # url(r'^blog/', include('blog.urls')),

    url(r'', include('cloudconfig.urls')),

)
