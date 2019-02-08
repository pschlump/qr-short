// URL to pull to get 2min hash:
// jQuery to make "get" call

// 1. Get the # put up and correct
// 2. Get the refresh button to work
// 3. Get the cowndown to work / black / red at 15 sec.
// 4. Some layout errors
// 5. get to work with 1 site first.


// example of sha256 hash
// var v = sha256("abc");
// console.log(v);

(function($) {
	'use strict';

	var app = {
		isLoading: true,
		visibleCards: {},
		selectedCities: [],
		spinner: document.querySelector('.loader'),
		cardTemplate: document.querySelector('.cardTemplate'),
		container: document.querySelector('.main'),
		addDialog: document.querySelector('.dialog-container'),
		daysOfWeek: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun']
	};

	// localStorage.setItem('2fa_hash', JSON.stringify(ls));
	var item_s = localStorage.getItem('2fa_hash');
	var item = JSON.parse(item_s);
	console.log ( 'ls data=', item);

	/*****************************************************************************
	 *
	 * Event listeners for UI elements
	 *
	 ****************************************************************************/

	$('#butRefresh').click( function() { // Refresh all of 2fa keys - pull new data from server.
		app.updateAllKeys();
	});

	$('#butAdd').click( function() { // Open/show the add new city dialog
		app.toggleAddDialog(true);
	});

	$('#butAddURL').click( function() {
		// Add the newly selected city
		var select = document.getElementById('selectCityToAdd');
		var selected = select.options[select.selectedIndex];
		var key = selected.value;
		var label = selected.textContent;
		if (!app.selectedCities) {
			app.selectedCities = [];
		}
		app.getForecast(key, label);
		app.selectedCities.push({key: key, label: label});
		app.saveSelectedCities();
		app.toggleAddDialog(false);
	});

	$('#butAddCancel').click( function() {	// Close Dialog
		app.toggleAddDialog(false);
	});


	/*****************************************************************************
	 *
	 * Methods to update/refresh the UI
	 *
	 ****************************************************************************/

	// Toggles the visibility of the add new city dialog.
	app.toggleAddDialog = function(visible) {
		if (visible) {
			app.addDialog.classList.add('dialog-container--visible');
		} else {
			app.addDialog.classList.remove('dialog-container--visible');
		}
	};

	// Updates a weather card with the latest weather forecast. If the card
	// doesn't already exist, it's cloned from the template.
	app.update2faData = function(data) {
		var dataLastUpdated = new Date(data.created);

		if ( 0 ) {
		var card = app.visibleCards[data.key];
		if (!card) {
			card = app.cardTemplate.cloneNode(true);
			card.classList.remove('cardTemplate');
			card.querySelector('.location').textContent = data.label;
			card.querySelector('.x2faValue').textContent = data.user1timeKey;
			card.querySelector('.countdown').textContent = data.ttlCurrent;
			card.removeAttribute('hidden');
			app.container.appendChild(card);
			app.visibleCards[data.key] = card;
		}

		// Verifies the data provide is newer than what's already visible
		// on the card, if it's not bail, if it is, continue and update the
		// time saved in the card
		var cardLastUpdatedElem = card.querySelector('.card-last-updated');
		var cardLastUpdated = cardLastUpdatedElem.textContent;
		if (cardLastUpdated) {
			cardLastUpdated = new Date(cardLastUpdated);
			// Bail if the card has more recent data then the data
			if (dataLastUpdated.getTime() < cardLastUpdated.getTime()) {
				return;
			}
		}
		var nextDays = card.querySelectorAll('.future .oneday');
		var today = new Date();
		today = today.getDay();
		} /* if ( 0 ) */
		if (app.isLoading) {
			app.spinner.setAttribute('hidden', true);
			app.container.removeAttribute('hidden');
			app.isLoading = false;
		}
	};


	/*****************************************************************************
	 *
	 * Methods for dealing with the model
	 *
	 ****************************************************************************/

	/*
	 * Gets a forecast for a specific city and updates the card with the data.
	 * getForecast() first checks if the weather data is in the cache. If so,
	 * then it gets that data and populates the card with the cached data.
	 * Then, getForecast() goes to the network for fresh data. If the network
	 * request goes through, then the card gets updated a second time with the
	 * freshest data.
	 */
	app.getForecast = function(key, label) {
		var statement = 'select * from weather.forecast where woeid=' + key;
		var url = 'https://query.yahooapis.com/v1/public/yql?format=json&q=' + statement;
		// TODO add cache logic here
		if ('caches' in window) {
			/*
			 * Check if the service worker has already cached this city's weather
			 * data. If the service worker has the data, then display the cached
			 * data while the app fetches the latest data.
			 */
			caches.match(url).then(function(response) {
				if (response) {
					response.json().then(function updateFromCache(json) {
						var results = json.query.results;
						results.key = key;
						results.label = label;
						// card.querySelector('.x2faValue').textContent = data.user1timeKey;
						// card.querySelector('.countdown').textContent = data.ttlCurrent;
						results.created = json.query.created;
						app.update2faData(results);
					});
				}
			});
		}
		// Fetch the latest data. -- change to use jQuery for ajax fetch
		/*
		var request = new XMLHttpRequest();
		request.onreadystatechange = function() {
			if (request.readyState === XMLHttpRequest.DONE) {
				if (request.status === 200) {
					var response = JSON.parse(request.response);
					var results = response.query.results;
					results.key = key;
					results.label = label;
					results.created = response.query.created;
					app.update2faData(results);
				}
			} else {
				// Return the initial weather forecast since no data is available.
				app.update2faData(initialWeatherForecast);
			}
		};
		request.open('GET', url);
		request.send();
		*/
		app.update2faData(initialWeatherForecast);
	};

	// Iterate all of the cards and attempt to get the latest forecast data
	// **done**
	app.updateAllKeys = function() {
		var keys = Object.keys(app.visibleCards);
		keys.forEach(function(key) {
			app.getForecast(key);
		});
	};

	// TODO add saveSelectedCities function here // Save list of cities to localStorage.
	app.saveSelectedCities = function() {
		var selectedCities = JSON.stringify(app.selectedCities);
		localStorage.selectedCities = selectedCities;
	};

	/*
	 * Fake weather data that is presented when the user first uses the app,
	 * or when the user has not saved any cities. See startup code for more
	 * discussion.
	 */
	var initialWeatherForecast = {
		key: '2459115',						// key needs to be a timestamp from 1970 - devided by 120 (timeout)
		label: 'http://127.0.0.1:9019', 	// the URL
		created: '2016-07-22T01:00:00Z',	
		// card.querySelector('.countdown').textContent = data.ttlCurrent;
		ttlCurrent: "120 sec",
		// card.querySelector('.x2faValue').textContent = data.user1timeKey;
		user1timeKey: "11331133",
		hash: "22322323232322323232323232323232323232323"
	};

	/************************************************************************
	 *
	 * Code required to start the app
	 *
	 * NOTE: To simplify this codelab, we've used localStorage.
	 *	 localStorage is a synchronous API and has serious performance
	 *	 implications. It should not be used in production applications!
	 *	 Instead, check out IDB (https://www.npmjs.com/package/idb) or
	 *	 SimpleDB (https://gist.github.com/inexorabletash/c8069c042b734519680c)
	 ************************************************************************/

	// TODO add startup code here
	app.selectedCities = localStorage.selectedCities;
	if (app.selectedCities) {
		app.selectedCities = JSON.parse(app.selectedCities);
		app.selectedCities.forEach(function(city) {
			app.getForecast(city.key, city.label);
		});
	} else {
		/* The user is using the app for the first time, or the user has not
		 * saved any cities, so show the user some fake data. A real app in this
		 * scenario could guess the user's location via IP lookup and then inject
		 * that data into the page.
		 */
		app.update2faData(initialWeatherForecast);
		app.selectedCities = [
			{key: initialWeatherForecast.key, label: initialWeatherForecast.label}
		];
		app.saveSelectedCities();
	}

//	// TODO add service worker code here
//	if ('serviceWorker' in navigator) {
//		navigator.serviceWorker
//						 .register('./service-worker.js')
//						 .then(function() { console.log('Service Worker Registered'); });
//	}

})(jQuery);
