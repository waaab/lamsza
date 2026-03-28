{
					window.__sveltekit_nicook = {
						base: ""
					};

					const element = document.body.firstElementChild;

					Promise.all([
						import("/app/immutable/entry/start.Dc8L_b5S.js"),
						import("/app/immutable/entry/app.Bwt5gKmQ.js")
					]).then(([kit, app]) => {
						kit.start(app, element);
					});
				}