/**
 * angular-ui-ace - This directive allows you to add ACE editor elements.
 * @version v0.1.1 - 2014-04-23
 * @link http://angular-ui.github.com
 * @license MIT
 */
"use strict";angular.module("ui.ace",[]).constant("uiAceConfig",{}).directive("uiAce",["uiAceConfig",function(a){if(angular.isUndefined(window.ace))throw new Error("ui-ace need ace to work... (o rly?)");var b=function(a,b,c){angular.isDefined(c.showGutter)&&a.renderer.setShowGutter(c.showGutter),angular.isDefined(c.useWrapMode)&&b.setUseWrapMode(c.useWrapMode),angular.isDefined(c.showInvisibles)&&a.renderer.setShowInvisibles(c.showInvisibles),angular.isDefined(c.showIndentGuides)&&a.renderer.setDisplayIndentGuides(c.showIndentGuides),angular.isDefined(c.useSoftTabs)&&b.setUseSoftTabs(c.useSoftTabs),angular.isDefined(c.disableSearch)&&c.disableSearch&&a.commands.addCommands([{name:"unfind",bindKey:{win:"Ctrl-F",mac:"Command-F"},exec:function(){return!1},readOnly:!0}]),angular.isFunction(c.onLoad)&&c.onLoad(a),angular.isString(c.theme)&&a.setTheme("ace/theme/"+c.theme),angular.isString(c.mode)&&b.setMode("ace/mode/"+c.mode)};return{restrict:"EA",require:"?ngModel",link:function(c,d,e,f){var g,h,i=a.ace||{},j=angular.extend({},i,c.$eval(e.uiAce)),k=window.ace.edit(d[0]),l=k.getSession(),m=function(){var a=arguments[0],b=Array.prototype.slice.call(arguments,1);angular.isDefined(a)&&c.$apply(function(){if(!angular.isFunction(a))throw new Error("ui-ace use a function as callback.");a(b)})},n={onChange:function(a){return function(b){var d=l.getValue();d===c.$eval(e.value)||c.$$phase||c.$root.$$phase||(angular.isDefined(f)&&c.$apply(function(){f.$setViewValue(d)}),m(a,b,k))}},onBlur:function(a){return function(){m(a,k)}}};e.$observe("readonly",function(a){k.setReadOnly("true"===a)}),angular.isDefined(f)&&(f.$formatters.push(function(a){if(angular.isUndefined(a)||null===a)return"";if(angular.isObject(a)||angular.isArray(a))throw new Error("ui-ace cannot use an object or an array as a model");return a}),f.$render=function(){l.setValue(f.$viewValue)}),b(k,l,j),c.$watch(e.uiAce,function(){j=angular.extend({},i,c.$eval(e.uiAce)),l.removeListener("change",g),g=n.onChange(j.onChange),l.on("change",g),k.removeListener("blur",h),h=n.onBlur(j.onBlur),k.on("blur",h),b(k,l,j)},!0),g=n.onChange(j.onChange),l.on("change",g),h=n.onBlur(j.onBlur),k.on("blur",h),d.on("$destroy",function(){k.session.$stopWorker(),k.destroy()}),c.$watch(function(){return[d[0].offsetWidth,d[0].offsetHeight]},function(){k.resize(),k.renderer.updateFull()},!0)}}}]);
