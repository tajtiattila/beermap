var map;

function initMap() {
  fetch("pubs.json")
    .then(function(response) {
      return response.json();
    })
    .then(function(mapData) {
      console.log(mapData.bounds);
      startMap(mapData);
    })
}

function startMap(mapData) {

  var styles = [
    {
        "featureType": "all",
        "elementType": "geometry.fill",
        "stylers": [
            {
                "weight": "2.00"
            }
        ]
    },
    {
        "featureType": "all",
        "elementType": "geometry.stroke",
        "stylers": [
            {
                "color": "#9c9c9c"
            }
        ]
    },
    {
        "featureType": "all",
        "elementType": "labels.text",
        "stylers": [
            {
                "visibility": "on"
            }
        ]
    },
    {
        "featureType": "landscape",
        "elementType": "all",
        "stylers": [
            {
                "color": "#f2f2f2"
            }
        ]
    },
    {
        "featureType": "landscape",
        "elementType": "geometry.fill",
        "stylers": [
            {
                "color": "#e8fce9"
            }
        ]
    },
    {
        "featureType": "landscape.man_made",
        "elementType": "geometry.fill",
        "stylers": [
            {
                "color": "#ffffff"
            }
        ]
    },
    {
        "featureType": "poi",
        "elementType": "all",
        "stylers": [
            {
                "visibility": "off"
            }
        ]
    },
    {
        "featureType": "road",
        "elementType": "all",
        "stylers": [
            {
                "saturation": -100
            },
            {
                "lightness": 45
            }
        ]
    },
    {
        "featureType": "road",
        "elementType": "geometry.fill",
        "stylers": [
            {
                "color": "#e1dfdf"
            }
        ]
    },
    {
        "featureType": "road",
        "elementType": "labels.text.fill",
        "stylers": [
            {
                "color": "#7b7b7b"
            }
        ]
    },
    {
        "featureType": "road",
        "elementType": "labels.text.stroke",
        "stylers": [
            {
                "color": "#ffffff"
            }
        ]
    },
    {
        "featureType": "road.highway",
        "elementType": "all",
        "stylers": [
            {
                "visibility": "simplified"
            }
        ]
    },
    {
        "featureType": "road.arterial",
        "elementType": "labels.icon",
        "stylers": [
            {
                "visibility": "off"
            }
        ]
    },
    {
        "featureType": "transit",
        "elementType": "all",
        "stylers": [
            {
                "saturation": "0"
            },
            {
                "lightness": "0"
            },
            {
                "visibility": "on"
            }
        ]
    },
    {
        "featureType": "transit",
        "elementType": "geometry",
        "stylers": [
            {
                "visibility": "on"
            },
            {
                "lightness": "0"
            }
        ]
    },
    {
        "featureType": "transit.station",
        "elementType": "all",
        "stylers": [
            {
                "saturation": "-42"
            },
            {
                "lightness": "0"
            },
            {
                "gamma": "1"
            }
        ]
    },
    {
        "featureType": "transit.station.bus",
        "elementType": "all",
        "stylers": [
            {
                "visibility": "off"
            }
        ]
    },
    {
        "featureType": "water",
        "elementType": "all",
        "stylers": [
            {
                "color": "#46bcec"
            },
            {
                "visibility": "on"
            }
        ]
    },
    {
        "featureType": "water",
        "elementType": "geometry.fill",
        "stylers": [
            {
                "color": "#c8d7d4"
            }
        ]
    },
    {
        "featureType": "water",
        "elementType": "labels.text.fill",
        "stylers": [
            {
                "color": "#070707"
            }
        ]
    },
    {
        "featureType": "water",
        "elementType": "labels.text.stroke",
        "stylers": [
            {
                "color": "#ffffff"
            }
        ]
    }
];

  var mapElement = document.getElementById('map');
  map = new google.maps.Map(mapElement, {
    styles: styles,
    mapTypeControl: false,
    scaleControl: true
  });
  map.fitBounds(mapData.bounds);

  var transitLayer = new google.maps.TransitLayer();
  transitLayer.setMap(map);

  var listControlDiv = document.createElement('div');
  listControlDiv.style.padding = "10px";
  var listControl = new ListControl(listControlDiv, map);

  listControlDiv.index = 1;
  map.controls[google.maps.ControlPosition.TOP_LEFT].push(listControlDiv);

  var infowindow = new google.maps.InfoWindow();

  var publist = document.getElementById("sidebar-content");
  mapData.pubs.forEach(function(p) {
    var marker = new google.maps.Marker({
        position: new google.maps.LatLng(p.lat, p.lng),
        icon: {
          url: p.icon,
          scaledSize: new google.maps.Size(28, 28)
        },
        map: map,
    });
    marker.addListener('click', function() {
      infowindow.setContent(p.content);
      infowindow.open(map, marker);
    });

    var div = document.createElement("div");
    div.className = "publist-item";
    var iconDiv = document.createElement("div");
    iconDiv.className = "publist-icon";
    var img = document.createElement("img");
    img.className = "publist-iconimg";
    img.src = p.icon;
    $(img).click(function() {
      infowindow.setContent(p.content);
      infowindow.open(map, marker);
    })
    iconDiv.appendChild(img);
    div.appendChild(iconDiv);
    var labelDiv = document.createElement("span");
    labelDiv.className = "publist-label";
    labelDiv.innerHTML = p.title;
    div.appendChild(labelDiv);
    publist.appendChild(div);
  });
}

function ListControl(controlDiv, map) {

  // Set CSS for the control border.
  var controlUI = document.createElement('div');
  controlUI.className = "mapcontrol-ui";
  controlUI.title = "Toggle pub list";
  controlDiv.appendChild(controlUI);

  // Set CSS for the control interior.
  var controlText = document.createElement('div');
  controlText.className = "mapcontrol-text";
  controlText.innerHTML = "List";
  controlUI.appendChild(controlText);

  // Setup the click event listeners: simply set the map to Chicago.
  controlUI.addEventListener('click', function() {
    var sidebarcontainer = document.getElementById("sidebarcontainer");
    if ($(sidebarcontainer).hasClass("hidden")) {
      $(sidebarcontainer).removeClass("hidden");
    } else {
      $(sidebarcontainer).addClass("hidden");
    }
    google.maps.event.trigger(map, "resize");
  });

}
