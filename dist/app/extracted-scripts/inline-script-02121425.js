{
					window.__sveltekit_1mfsimu = {
						base: ""
					};

					const element = document.body.firstElementChild;

					Promise.all([
						import("/app/immutable/entry/start.CnbhRQbg.js"),
						import("/app/immutable/entry/app.Cc2RRRA6.js")
					]).then(([kit, app]) => {
						kit.start(app, element);
					});
				}