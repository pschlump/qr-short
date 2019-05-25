
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

