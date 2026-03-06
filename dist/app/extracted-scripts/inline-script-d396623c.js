{
					window.__sveltekit_eqstvc = {
						base: ""
					};

					const element = document.body.firstElementChild;

					Promise.all([
						import("/app/immutable/entry/start.Dfm63XyL.js"),
						import("/app/immutable/entry/app.BzYTpxK1.js")
					]).then(([kit, app]) => {
						kit.start(app, element);
					});
				}