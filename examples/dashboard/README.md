## api dashboard

![dash](assets/dash.jpg)

### usage
1. Setup ENV Vars
    - rename `.env.sample` to `.env`
    - setup the influxdb user and pass in `.env`
    - if you want to change the grafana user details then go to `grafana/config/grafana.ini`
2. `docker-compose up --build`
3. Go to `localhost:3000`, login with your grafana details and from the dashboards folder, open the 4chan stats dashboard

voila!
