function fnShowModal(ip) {
    $.ajax({
        url: "/ipcount/whois/" + ip,
        success: function(result) {
            $("#modal-paragraph").html("<pre>" + result.d + "</pre>")
            $("#whoisModal").modal({
                keyboard: true,
            })
        }
    })
    e.preventDefault();
}

$(document).ready(function() {
    $('#scores5m').on('click', 'td', function(e) {
        var ip = $(this).html()
        fnShowModal(ip)
    });

    $('#scores1h').on('click', 'td', function(e) {
        var ip = $(this).html()
        fnShowModal(ip)
    })

    $('#scores12h').on('click', 'td', function(e) {
        var ip = $(this).html()
        fnShowModal(ip)
    })

    $('#scores24h').on('click', 'td', function(e) {
        var ip = $(this).html()
        fnShowModal(ip)
    })

    var scoresFive = $('#scores5m').dataTable({
        "bProcessing": true,
        "sAjaxSource": '/ipcount/top/1',
        "sAjaxDataProp": "d",
        "sPaginationType": "two_button",
        "aaSorting": [
            [1, "desc"]
        ]
    });

    $('#scores1h').dataTable({
        "bProcessing": true,
        "sAjaxSource": '/ipcount/top/2',
        "sAjaxDataProp": "d",
        "sPaginationType": "two_button",
        "aaSorting": [
            [1, "desc"]
        ],
    });

    $('#scores12h').dataTable({
        "bProcessing": true,
        "sAjaxSource": '/ipcount/top/3',
        "sAjaxDataProp": "d",
        "sPaginationType": "two_button",
        "aaSorting": [
            [1, "desc"]
        ],
    });

    $('#scores24h').dataTable({
        "bProcessing": true,
        "sAjaxSource": '/ipcount/top/4',
        "sAjaxDataProp": "d",
        "sPaginationType": "two_button",
        "aaSorting": [
            [1, "desc"]
        ],
    });

});
