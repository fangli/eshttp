'use strict';


angular.module('eshttp')
  .controller('HeadCtrl', function ($scope, $rootScope, $location) {
    $rootScope.menus = [
      ['Nodes', '/node'],
      ['Configuration', '/config'],
      ['Monitoring', '/status'],
    ];

    $scope.isActive = function (viewLocation) {
      return $location.path().startsWith(viewLocation);
    };

  });
