'use strict';

angular.module('eshttp')
  .filter('unixtostr', function() {
    return function(str){
        var dt;
        dt = Date.create(str * 1000).format('{yyyy}-{MM}-{dd} {HH}:{mm}:{ss}');
        if (dt == "Invalid Date") {
            return 'N/A';
        } else {
            return dt;
        }
    };
  });
