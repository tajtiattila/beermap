var map;

function initMap() {
  fetch("mapstyle.json")
    .then(function(response) {
      return response.json();
    })
    .then(function(mapStyle) {
      fetch("pubs.json")
        .then(function(response) {
          return response.json();
        })
        .then(function(mapData) {
          startMap(mapStyle, mapData);
        })
    })
}

function startMap(styles, mapData) {

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
