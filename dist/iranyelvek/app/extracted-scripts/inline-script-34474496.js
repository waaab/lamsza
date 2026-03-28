(function () {
			var t = localStorage.getItem('theme');
			if (t === 'dark' || t === 'light') {
				document.documentElement.setAttribute('data-theme', t);
			}
			// 'system' or missing: leave data-theme unset → CSS media query handles it
		})();