angular.module('gobotApp', [])
.controller('GameController', function($scope) {
  var game = this;
  var playback = Gobot.GetPlayback(replayStr);
  game.round = 0;

  game.updateBoard = function(board) {
    game.board = board;
    game.rows = new Array(board.Height())
    for (var y = 0; y < board.Height(); y++) {
      game.rows[y] = new Array(board.Width())
      for (var x = 0; x < board.Width(); x++) {
        game.rows[y][x] = board.AtXY(x,y)
      }
    }
    $scope.$apply();
  }

  //game.updateBoard(playback.Board(0));
  var id = window.setInterval(function() {
    game.round++;
    var board = playback.Board(game.round);
    game.updateBoard(board);
    if (game.round >= 99) {
      window.clearInterval(id)
    }
  }, 200);
})
.config(function($interpolateProvider) {
  $interpolateProvider.startSymbol('[[');
  $interpolateProvider.endSymbol(']]');
});
