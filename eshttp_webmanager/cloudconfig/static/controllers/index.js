'use strict';

/**
 * @ngdoc function
 * @name crsApp.controller:ProfileCtrl
 * @description
 * # ProfileCtrl
 * Controller of the crsApp
 */
angular.module('eshttp')
  .controller('IndexCtrl', function ($scope, $filter, Api) {

    $scope.loadNodes = function() {
      $scope.isLoading = true;
      $scope.nodes = [];
      Api.get('/api/node')
        .success(function(data){
          $scope.isLoading = false;
          if (data.result == 'ok') {
            $scope.nodes = data.body;
          } else {
            window.alert(data.message);
          }
        })
        .error(function(){
          $scope.isLoading = false;
          window.alert('We can not load nodes currently, please refresh the page!');
        });
    };

    $scope.loadConfigs = function() {
      $scope.configs = [];
      Api.get('/api/config')
        .success(function(data){
          if (data.result == 'ok') {
            $scope.configs = data.body;
          } else {
            window.alert(data.message);
          }
        })
        .error(function(){
          window.alert('We can not load nodes currently, please refresh the page!');
        });
    };

    $scope.saveNode = function(node) {

      if (!node.config_id) {
        window.alert('You must set a valid config template for this node');
        return;
      }

      node.isSaving = true;
      Api.post('/api/node/edit', node)
        .success(function(data){
          node.isSaving = false;
          if (data.result == 'ok') {
            node.config_name = $filter('filter')($scope.configs, {'id': node.config_id})[0].name;
            node.isEditing = false;
          } else {
            window.alert(data.message);
          }
        })
        .error(function(){
          node.isSaving = false;
          window.alert('We can not save node currently, please refresh the page!');
        });
    };

    $scope.confirmNode = function(node) {
      node.isSaving = true;
      Api.post('/api/node/confirm', node)
        .success(function(data){
          node.isSaving = false;
          if (data.result == 'ok') {
            node.confirmed = true;
          } else {
            window.alert(data.message);
          }
        })
        .error(function(){
          node.isSaving = false;
          window.alert('We can not add this node currently, please refresh the page!');
        });
    };

    $scope.deleteNode = function(node) {
      node.isSaving = true;
      Api.post('/api/node/delete', node)
        .success(function(data){
          node.isSaving = false;
          if (data.result == 'ok') {
            for (var i = $scope.nodes.length - 1; i >= 0; i--) {
              if ($scope.nodes[i].id === data.body.id) {
                $scope.nodes.splice(i, 1);
              }
            }
          } else {
            window.alert(data.message);
          }
        })
        .error(function(){
          node.isSaving = false;
          window.alert('We can not add this node currently, please refresh the page!');
        });
    };

    $scope.initial = function() {
        $scope.nodes = [];
        $scope.configs = [];
        $scope.loadNodes();
        $scope.loadConfigs();
    };

    $scope.initial();

  });
