'use strict';

/**
 * @ngdoc overview
 * @name crsApp
 * @description
 * # crsApp
 *
 * Main module of the application.
 */
angular
  .module('eshttp', [
    'ngRoute',
    'ui.ace',
  ])
  .config(function ($routeProvider) {
    $routeProvider
      .when('/node', {
        templateUrl: '/static/views/index.html',
        controller: 'IndexCtrl'
      })
      .when('/config', {
        templateUrl: '/static/views/config.html',
        controller: 'ConfigCtrl'
      })
      .when('/status', {
        templateUrl: '/static/views/status.html',
        controller: 'StatusCtrl'
      })

      .otherwise({
        redirectTo: '/node'
      });
  });
