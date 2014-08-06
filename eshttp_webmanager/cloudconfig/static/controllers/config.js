'use strict';

/**
 * @ngdoc function
 * @name crsApp.controller:ProfileCtrl
 * @description
 * # ProfileCtrl
 * Controller of the crsApp
 */
angular.module('eshttp')
  .controller('ConfigCtrl', function ($scope, $filter, Api) {

    $scope.loadConfigs = function() {
      $scope.isLoading = true;
      $scope.configs = [];
      Api.get('/api/config')
        .success(function(data){
          $scope.isLoading = false;
          if (data.result == 'ok') {
            $scope.configs = data.body;
          } else {
            window.alert(data.message);
          }
        })
        .error(function(){
          $scope.isLoading = false;
          window.alert('We can not load configs currently, please refresh the page!');
        });
    };

    $scope.saveConfig = function(config) {
      config.isSaving = true;
      Api.post('/api/config/edit', config)
        .success(function(data){
          config.isSaving = false;
          if (data.result == 'ok') {
            config.isEditing = false;
          } else {
            window.alert(data.message);
          }
        })
        .error(function(){
          config.isSaving = false;
          window.alert('We can not save config currently, please refresh the page!');
        });
    };

    $scope.addConfig = function(config) {
      config.isSaving = true;
      Api.post('/api/config/add', config)
        .success(function(data){
          config.isSaving = false;
          if (data.result == 'ok') {
            config.name='';
            config.content='';
            $scope.loadConfigs();
            $scope.showingNew = false;
          } else {
            window.alert(data.message);
          }
        })
        .error(function(){
          config.isSaving = false;
          window.alert('We can not add config currently, please refresh the page!');
        });
    };

    $scope.deleteConfig = function(config) {
      config.isSaving = true;
      Api.post('/api/config/delete', config)
        .success(function(data){
          config.isSaving = false;
          if (data.result == 'ok') {
            for (var i = $scope.configs.length - 1; i >= 0; i--) {
              if ($scope.configs[i].id === data.body.id) {
                $scope.configs.splice(i, 1);
              }
            }
          } else {
            window.alert(data.message);
          }
        })
        .error(function(){
          config.isSaving = false;
          window.alert('We can not delete this config currently, please refresh the page!');
        });
    };

    $scope.aceInitial = function() {
      $scope.ace.option = {
        theme: 'monokai',
        maxLines: 'Infinity',
        mode: 'ini',
      };
    };

    $scope.initial = function() {
        $scope.ace = {};
        $scope.configs = [];
        $scope.loadConfigs();
        $scope.aceInitial();
    };

    $scope.initial();

  });
