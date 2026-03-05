{
					window.__sveltekit_1mfsimu = {
						base: new URL(".", location).pathname.slice(0, -1)
					};

					const element = document.body.firstElementChild;

					Promise.all([
						import("../../app/immutable/entry/start.CnbhRQbg.js"),
						import("../../app/immutable/entry/app.Cc2RRRA6.js")
					]).then(([kit, app]) => {
						kit.start(app, element, {
							node_ids: [0, 3, 17],
							data: [null,null,null],
							form: null,
							error: null
						});
					});
				}