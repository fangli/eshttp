'use strict';


angular.module('eshttp')
  .factory('Api', ['$http', function($http) {
    var baseUrl = '';
    return {
      get: function(uri) {
        return $http.get(baseUrl + uri);
      },
      post: function(uri, data){
        return $http.post(baseUrl + uri, data);
      },
    };
  }]);
