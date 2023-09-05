watcher := go run github.com/barelyhuman/gomon

devServer:
	$(watcher) -w "." -exclude="storage.json" .

devStyles:
	npx tailwindcss -i ./assets/main.css -o ./dist/style.css --watch
dev:
	${MAKE} -j4 devStyles devServer