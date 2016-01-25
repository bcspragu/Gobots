angular.module('gobotApp', [])
.controller('GameController', function() {
  var game = this;
  game.round = 0;

  game.setBoard = function(board) {
    game.board = board
    game.rows = new Array(board.Height())
    for (var y = 0; y < board.Height(); y++) {
      game.rows[y] = new Array(board.Width())
      for (var x = 0; x < board.Width(); x++) {
        game.rows[y][x] = board.AtXY(x,y)
      }
    }
  }

  $.get(buildUrl(game.round), function(resp) {
    game.board = Gobot.GetBoard(resp)
  });
  window.setTimeout(function() {
    game.round++;
    $.get(buildUrl(game.round), function(resp) {
      game.board = Gobot.GetBoard(resp)
    });
  }, 1000);
})
.config(function($interpolateProvider) {
  $interpolateProvider.startSymbol('[[');
  $interpolateProvider.endSymbol(']]');
});

function buildUrl(round) {
  return "http://" + Host + "/game/" + GameID + "/" + round
}
