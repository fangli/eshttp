'use strict';

angular.module('eshttp')
  .filter('humanbytes', function() {
    return function(bytes) {
        if (bytes == 0){
            return '0 byte';
        }

        if (bytes < 0 ){
            return 'N/A byte';
        }

        if (!((bytes - 0) == bytes && (''+bytes).replace(/^\s+|\s+$/g, "").length > 0)){
            return 'N/A byte'
        }

        var s = ['bytes', 'kB', 'MB', 'GB', 'TB', 'PB'];
        var e = Math.floor(Math.log(bytes) / Math.log(1024));
        return (bytes / Math.pow(1024, e)).toFixed(2) + " " + s[e];
    };
  });


