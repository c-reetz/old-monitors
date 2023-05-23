#  Gamle monitors
Brugte følgende monitors til at holde øje med price-errors på Ralph-Lauren, samt til at monitor specificerede produkter på Unisport DK. Koden er slet ikke pæn, men den virkede og var hurtigere end de andre på markedet. Brug den som inspiration, eller til at holde øje med price errors.

# Selve monitorerne
Connectede til MongoDB, hentede en til flere webhooks derfra - til RL var det vigtigt at have flere, da der tit blev sendt flere webhooks af gangen, så jo flere = mindre ratelimiting = hurtigere notifikation. Gemte også produkter i DB samt data om dem, så man nemt kunne genstarte monitors og de ikke skulle til at gendanne cache på hvad der var sendt allerede.
