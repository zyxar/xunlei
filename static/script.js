;(function(){

$(window).ready(function() {
  $.ajax({
    type: 'get',
    url: 'http://localhost:8808/tasks/raw/1'
  }).done(function(resp){
    if (typeof resp === 'string') resp = JSON.parse(resp);
    var tasks = resp.info.tasks;
    for (var i in tasks) {
      var container = $('<div></div>').attr('id', tasks[i].id);
      var element = $('<a></a>').text(tasks[i].id+' '+tasks[i].taskname).attr('href', '#');
      $(container).append(element);
      $(element).click(function (evt) {
        console.log(tasks[i]);
      });
      $('#tasks').append(container);
    }
  });

  // setTimeout(function(){
  //   $.ajax({
  //     type: 'POST',
  //     url: '/task',
  //     data: JSON.stringify ({
  //       action: 'add',
  //       data: "https://pqrs.org/macosx/keyremap4macbook/files/KeyRemap4MacBook-9.3.0.dmg"
  //     }),
  //     contentType: "application/json",
  //     dataType: 'json'
  //   }).done(function(data) {
  //     console.log(data);
  //   });
  // }, 1000);

});

}());
