'use strict';

/**
 * @ngdoc function
 * @name crsApp.controller:ProfileCtrl
 * @description
 * # ProfileCtrl
 * Controller of the crsApp
 */
angular.module('eshttp')
  .controller('StatusCtrl', function ($scope, $filter, $timeout, Api) {

    $scope.frefreshTimout = undefined;

    $scope.decodeExtrem = function(lst) {
      var minStarted = 99999999999999;
      var minSize = 0;
      var maxEnded = 0;
      var maxSize = 0;
      for (var i = lst.length - 1; i >= 0; i--) {
        if (lst[i] === null) {
          continue;
        }
        minStarted = Math.min(lst[i].started, minStarted);
        maxEnded = Math.max(lst[i].started+lst[i].took, maxEnded);
        minSize = Math.min(lst[i].size, minSize);
        maxSize = Math.max(lst[i].size, maxSize);
      }
      return {
        begin: minStarted,
        end: maxEnded,
        max: maxSize,
        min: minSize,
      };
    };

    $scope.mergeNodes = function(nodes) {
      var mergedS3 = [];
      var mergedEs = [];
      for (var i = nodes.length - 1; i >= 0; i--) {
        if (nodes[i].status.sender) {
          mergedS3 = mergedS3.concat(mergedS3, nodes[i].status.sender.s3);
          mergedEs = mergedEs.concat(mergedEs, nodes[i].status.sender.es);
        }
      }
      return {
        s3: mergedS3,
        es: mergedEs,
      };
    };

    $scope.avgCalc = function(node) {
      if (!node.status.sender) {
        return {
          s3: 0,
          es: 0,
        };
      }

      var s3Total = 0;
      var s3Time = 0;
      var esTotal = 0;
      var esTime = 0;
      var s3Count = 0;
      var esCount = 0;
      for (var i = node.status.sender.s3.length - 1; i >= 0; i--) {
        if (node.status.sender.s3[i]){
          s3Total += node.status.sender.s3[i].size;
          s3Time += node.status.sender.s3[i].took;
          s3Count ++;
        }
      }

      for (var j = node.status.sender.es.length - 1; j>= 0; j--) {
        if (node.status.sender.es[j]){
          esTotal += node.status.sender.es[j].size;
          esTime += node.status.sender.es[j].took;
          esCount ++;
        }
      }
      return {
        s3Byte: s3Total/s3Time,
        esByte: esTotal/esTime,
        s3Time: s3Time/s3Count,
        esTime: esTime/esCount,
      };
    };

    $scope.decode = function(nodes) {
      var merged = $scope.mergeNodes(nodes);
      var extremS3 = $scope.decodeExtrem(merged.s3);
      var extremEs = $scope.decodeExtrem(merged.es);
      var timelengthS3 = extremS3.end - extremS3.begin;
      var timelengthEs = extremEs.end - extremEs.begin;

      var sizelengthS3 = extremS3.max - extremS3.min;
      var sizelengthEs = extremEs.max - extremEs.min;

      var s3Formated = [];

      for (var i = nodes.length - 1; i >= 0; i--) {

        if (!nodes[i].status.sender) {
          continue;
        }

        nodes[i].status.sender.s3Formated = [];
        nodes[i].status.sender.esFormated = [];

        nodes[i].status.sender.avg = $scope.avgCalc(nodes[i]);

        for (var j = 0; j < nodes[i].status.sender.s3.length; j++) {
          if (!nodes[i].status.sender.s3[j]) {
            continue;
          }
          nodes[i].status.sender.s3Formated.push({
            marginLeft: (nodes[i].status.sender.s3[j].started - extremS3.begin)*100/timelengthS3,
            len: nodes[i].status.sender.s3[j].took*100/timelengthS3,
            size: nodes[i].status.sender.s3[j].size/extremS3.max*5,
            result: nodes[i].status.sender.s3[j].result,
          });
        }

        for (var k = 0; k < nodes[i].status.sender.es.length; k++) {
          if (!nodes[i].status.sender.es[k]) {
            continue;
          }
          nodes[i].status.sender.esFormated.push({
            marginLeft: (nodes[i].status.sender.es[k].started - extremEs.begin)*100/timelengthEs,
            len: nodes[i].status.sender.es[k].took*100/timelengthEs,
            size: nodes[i].status.sender.es[k].size/extremEs.max*5,
            result: nodes[i].status.sender.es[k].result,
          });
        }

      }
    };

    $scope.loadStatus = function() {
      $scope.isLoading = true;
      Api.get('/api/status')
        .success(function(data){
          $scope.isLoading = false;
          if (data.result == 'ok') {
            $scope.decode(data.body);
            $scope.status = data.body;
            if ($scope.frefreshTimout) {
              $timeout.cancel($scope.frefreshTimout);
            }
            $scope.frefreshTimout = $timeout($scope.loadStatus, $scope.refreshInterval*1000);
          } else {
            window.alert(data.message);
          }
        })
        .error(function(){
          $scope.isLoading = false;
          window.alert('We can not load status currently, please refresh the page!');
        });
    };

    $scope.initial = function() {
        $scope.status = [];
        $scope.loadStatus();
        $scope.refreshInterval = 1;
    };

    $scope.initial();

  });
