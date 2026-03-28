{
					window.__sveltekit_nicook = {
						base: new URL(".", location).pathname.slice(0, -1)
					};

					const element = document.body.firstElementChild;

					Promise.all([
						import("../../app/immutable/entry/start.Dc8L_b5S.js"),
						import("../../app/immutable/entry/app.Bwt5gKmQ.js")
					]).then(([kit, app]) => {
						kit.start(app, element, {
							node_ids: [0, 2, 10],
							data: [null,null,null],
							form: null,
							error: null
						});
					});
				}