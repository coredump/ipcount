function fnGetSelected(oTableLocal) {
  return oTableLocal.$('tr.row_selected');
}

$(document).ready(function() {

  $('#scores5m').on('click', 'td', function(e) {
    var ip = $(this).html()
    $.ajax({
      url: "/whois/" + ip,
      success: function(result) {
        console.log(result.d)
      }
    })
    e.preventDefault();
  });

  var scoresFive = $('#scores5m').dataTable({
    "bProcessing": true,
    "sAjaxSource": '/top/1',
    "sAjaxDataProp": "d",
    "sPaginationType": "two_button",
    "aaSorting": [
      [1, "desc"]
    ]
  });

  $('#scores1h').dataTable({
    "bProcessing": true,
    "sAjaxSource": '/top/2',
    "sAjaxDataProp": "d",
    "sPaginationType": "two_button",
    "aaSorting": [
      [1, "desc"]
    ],
  });

  $('#scores12h').dataTable({
    "bProcessing": true,
    "sAjaxSource": '/top/3',
    "sAjaxDataProp": "d",
    "sPaginationType": "two_button",
    "aaSorting": [
      [1, "desc"]
    ],
  });

  $('#scores24h').dataTable({
    "bProcessing": true,
    "sAjaxSource": '/top/4',
    "sAjaxDataProp": "d",
    "sPaginationType": "two_button",
    "aaSorting": [
      [1, "desc"]
    ],
  });



});
