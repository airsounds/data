trigger-fetch:
	@curl -XPOST \
		-H "Authorization: token ${GITHUB_TOKEN}" \
		-d '{"event_type":"open"}' \
		https://api.github.com/repos/airsounds/data/dispatches