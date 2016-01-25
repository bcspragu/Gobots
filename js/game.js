angular.module('gobotApp', [])
.controller('GameController', function($scope) {
  var game = this;
  game.round = 0;

  game.updateBoard = function(board) {
    game.board = board
    game.rows = new Array(board.Height())
    for (var y = 0; y < board.Height(); y++) {
      game.rows[y] = new Array(board.Width())
      for (var x = 0; x < board.Width(); x++) {
        game.rows[y][x] = board.AtXY(x,y)
      }
    }
    $scope.$apply();
  }

  $.get(buildUrl(game.round), function(resp) {
    game.updateBoard(Gobot.GetBoard(resp));
  });
  var id = window.setInterval(function() {
    game.round++;
    $.get(buildUrl(game.round), function(resp) {
      game.updateBoard(Gobot.GetBoard(resp));
    });
    if (game.round >= 99) {
      window.clearInterval(id)
    }
  }, 1000);
})
.config(function($interpolateProvider) {
  $interpolateProvider.startSymbol('[[');
  $interpolateProvider.endSymbol(']]');
});

function buildUrl(round) {
  return "http://" + Host + "/game/" + GameID + "/" + round
}
