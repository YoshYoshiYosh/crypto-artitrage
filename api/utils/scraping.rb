require 'net/http'
require 'nokogiri'

# htmlの取得
url = "https://info.finance.yahoo.co.jp/fx/detail/?code=usdjpy"
uri = URI(url)
html = Net::HTTP.get(uri)

# htmlの変換
doc = Nokogiri::HTML.parse(html, nil, 'utf-8')

# 底値の取得
low_nodes = doc.css('dd#USDJPY_detail_low').text.to_f
# 高値の取得
heigh_nodes = doc.css('dd#USDJPY_detail_high').text.to_f

print ((low_nodes + heigh_nodes)/2).round(4)


