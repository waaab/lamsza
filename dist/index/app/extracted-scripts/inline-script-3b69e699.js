{
					window.__sveltekit_eqstvc = {
						base: new URL("..", location).pathname.slice(0, -1)
					};

					const element = document.body.firstElementChild;

					Promise.all([
						import("../../../app/immutable/entry/start.Dfm63XyL.js"),
						import("../../../app/immutable/entry/app.BzYTpxK1.js")
					]).then(([kit, app]) => {
						kit.start(app, element, {
							node_ids: [0, 2, 12],
							data: [null,null,null],
							form: null,
							error: null
						});
					});
				}