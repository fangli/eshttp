#!/usr/bin/env python
# -*- coding:utf-8 -*-

from django.shortcuts import render

def index(r):
    return render(r, 'index.html')
