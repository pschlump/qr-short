
function createCookie(name,value,days) {
	var expires = "";
	if (days) {
		var date = new Date();
		date.setTime(date.getTime()+(days*24*60*60*1000));
		expires = "; expires="+date.toGMTString();
	} 
	document.cookie = name+"="+value+expires+"; path=/";
}

function getCookie(name) {
	var nameEQ = name + "=";
	var ca = document.cookie.split(';');
	for(var i=0;i < ca.length;i++) {
		var c = ca[i];
		while (c.charAt(0)==' ') c = c.substring(1,c.length);
		if (c.indexOf(nameEQ) == 0) {
			return c.substring(nameEQ.length,c.length);
		}
	}
	return null;
}

function delCookie(name) {
	createCookie(name,"",-1);
}

function URLStrToHash(query) {
	var rv = {};
	var decode = function (s) { return decodeURIComponent(s.replace(/^\?/,"").replace(/\+/g, " ")); };

	var p1 = query.replace(/([^&=]+)=?([^&]*)/g,function(j,n,v){
		n = decode(n);
		v = decode(v);
		if ( typeof(rv[n]) === "undefined" ) {
			rv[n] = ( typeof v === "undefined" ) ? "" : v;
		} else if ( typeof(rv[n]) === "string" ) {
			var x = rv[n];
			rv[n] = [];
			rv[n].push ( x );
			rv[n].push ( v );
		} else {
			rv[n].push ( v );
		}
		return "";
	});
	return rv;
}

//var v = URLStrToHash("a=12&b=22&c&d");
//console.log ( 'v=', v );
//var v = URLStrToHash("a=12&b=22&x=aa&x=bb&x=cc&x=dd&x=ee&d=888888");
//console.log ( 'v=', v );

function Id(x){
	return document.getElementById(x);
}

var g_origin = window.location.origin;
if ( ! g_origin ) {			// Probablyl running on Opera
	g_origin = window.location.protocol + "//" + window.location.host;
}

var g_param = URLStrToHash ( window.location.search );

console.log ( window.location.search );
console.log ( "g_param=", g_param );

var rc = getCookie('recovery_token');
if ( rc != "" ) {
	console.log ( "Cooky recovery token set",rc);
}
if ( g_param.recovery_token ) {
	console.log ( "Setting recovery token",g_param.recovery_token);
	createCookie('recovery_token',g_param.recovery_token);
}
