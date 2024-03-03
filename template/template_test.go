package template

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestCompileTemplate(t *testing.T) {
	body := `
<p dir="ltr"><span>Thanh test</span></p>
<p dir="ltr"><span>Mot hai ban nam sau </span><span class="sbz-dynamic-field" data-dynamic-field="user.fullname">Họ tên</span></p>
<p dir="ltr"><span class="sbz-dynamic-field" data-dynamic-field="user.emails">Email</span></p>
<p dir="ltr"><a href="https://google.com" target="_blank" rel="noopener noreferrer"><span>Google</span></a></p>
<p><span>--</span></p>
<p dir="ltr"><b><strong class="sbz_lexical_text__bold">BBBBBBBBB</strong></b></p>
<p dir="ltr"><i><em class="sbz_lexical_text__italic">IIIIIIIIIII</em></i></p>
<p dir="ltr"><u><span class="sbz_lexical_text__underline">UUUUUUUUUUUU</span></u></p>
<p><span>--</span></p>
<p dir="ltr" style="text-align: center;"><span>center</span></p>
<p><br></p>
`
	out := CompileTemplateDynamicField(body, map[string]string{"user.fullname": "Thanh", "user.emails": "thanhpk@live.<script>com"})
	fmt.Println("OUT", out)
}

func TestToTextPlain(t *testing.T) {
	testCases := []struct {
		in  string
		out string
	}{{
		in: "<p class=\"sbz_lexical_paragraph\" dir=\"ltr\"><span>XIn chao ban </span><span class=\"sbz-dynamic-field\" data-dynamic-field=\"user.fullname\">Tên khách</span><br><span class=\"lexical-emoji neutral\"><span class=\"lexical-emoji-inner\">😐</span></span></p>",
		out: `XIn chao ban Tên khách
😐`,
	}, {
		in:  "<p class=\"sbz_lexical_paragraph\" dir=\"ltr\"><span> Hiện tại hôm nay đang có chương trình ưu đãi đặc biệt dành cho que 4 Femplant chị nhé. Hôm nay đặt hẹn chị chỉ còn trả cho que cấy 4 năm Femplant là : 1.000.000 + 150.000 Phí khám tư vấn + tặng các dịch vụ khám sàng lọc ( chưa bao gồm thuốc sau cấy tùy vào cơ địa mỗi người bác sĩ sẽ kê đơn) . Chị có muốn đăng kí nhận ưu đãi này không ạ?</span></p>",
		out: "Hiện tại hôm nay đang có chương trình ưu đãi đặc biệt dành cho que 4 Femplant chị nhé. Hôm nay đặt hẹn chị chỉ còn trả cho que cấy 4 năm Femplant là : 1.000.000 + 150.000 Phí khám tư vấn + tặng các dịch vụ khám sàng lọc ( chưa bao gồm thuốc sau cấy tùy vào cơ địa mỗi người bác sĩ sẽ kê đơn) . Chị có muốn đăng kí nhận ưu đãi này không ạ?",
	}, {
		in: `<p>xin chào bạn. Dưới đây là một số điểm bạn cần lưu ý</p><ul><li>Đi làm đúng giờ theo quy định ở <a href="https://subiz.com.vn">đây</a></li>
<li>Không ăn quà vặt</li></ul><span>Và đây là dấu xuống&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;<b>dòng:</b><br/>W&nbsp;ANT</span>
<script>topci</script>
One two three four five
<b>huh</b>
<!-- commentno -->
<script type='text/javascript'>
<a>stupid</a>
/* <![CDATA[ */
var post_notif_widget_ajax_obj = {"ajax_url":"http:\/\/site.com\/wp-admin\/admin-ajax.php","nonce":"9b8270e2ef","processing_msg":"Processing..."};
/* ]]> */
</script>`,
		out: `xin chào bạn. Dưới đây là một số điểm bạn cần lưu ýĐi làm đúng giờ theo quy định ở đây
Không ăn quà vặtVà đây là dấu xuống     dòng:
W ANT

One two three four five
huh`,
	}, {
		in: "\u003cp class=\"sbz_lexical_paragraph\" dir=\"ltr\"\u003e\u003cspan style=\"white-space: pre-wrap;\"\u003eSGO DMC mến chào anh/chị \u003c/span\u003e\u003cspan class=\"sbz-dynamic-field\" data-dynamic-field=\"user.name\" style=\"white-space: pre-wrap;\"\u003eTên khách\u003c/span\u003e\u003c/p\u003e\u003cp class=\"sbz_lexical_paragraph\" dir=\"ltr\"\u003e\u003cspan style=\"white-space: pre-wrap;\"\u003eTour trải nghiệm đặc biệt tham gia giải chạy Marathon Gyeongju hoa anh đào được tài trợ BIB chạy bởi Tổng cục du lịch Hàn Quốc khởi hành duy nhất ngày 5/4.\u003c/span\u003e\u003c/p\u003e\u003cp class=\"sbz_lexical_paragraph\" dir=\"ltr\"\u003e\u003cbr\u003e\u003cspan style=\"white-space: pre-wrap;\"\u003eAnh/chị dự định đăng ký tham gia mấy thành viên ạ? Em xin thông tin để có thể hỗ trợ mình chi tiết ạ.\u003c/span\u003e\u003c/p\u003e",
		out: `SGO DMC mến chào anh/chị Tên khách
Tour trải nghiệm đặc biệt tham gia giải chạy Marathon Gyeongju hoa anh đào được tài trợ BIB chạy bởi Tổng cục du lịch Hàn Quốc khởi hành duy nhất ngày 5/4.

Anh/chị dự định đăng ký tham gia mấy thành viên ạ? Em xin thông tin để có thể hỗ trợ mình chi tiết ạ.`,
	}}
	for _, tc := range testCases {
		actual := CompileTemplateToPlainText(tc.in)
		if actual != tc.out {
			jsonactual, _ := json.Marshal(actual)
			jsonout, _ := json.Marshal(tc.out)
			t.Error("SHOULD BE EQ", string(jsonactual), "|GOT|", string(jsonout))
		}
	}
}

func TestCompileTemplateToEmail(t *testing.T) {
	body := `<p dir="ltr"><span>Thanh test</span></p>
<p dir="ltr"><span>Mot hai ban nam sau </span><span class="sbz-dynamic-field" data-dynamic-field="user.fullname">Họ tên</span></p>
<p dir="ltr"><span class="sbz-dynamic-field" data-dynamic-field="user.emails">Email</span></p>
<p dir="ltr"><a href="https://google.com" target="_blank" rel="noopener noreferrer"><span>Google</span></a></p>
<p><span>--</span></p>
<p dir="ltr"><b><strong class="sbz_lexical_text__bold">BBBBBBBBB</strong></b></p>
<p dir="ltr"><i><em class="sbz_lexical_text__italic">IIIIIIIIIII</em></i></p>
<p dir="ltr"><u><span class="sbz_lexical_text__underline">UUUUUUUUUUUU</span></u></p>
<p><span>--</span></p>
<p dir="ltr" style="text-align: center;"><span>center</span></p>
<p><br></p>`

	out := CompileTemplateToEmail(body, map[string]string{"user.fullname": "Thanh", "user.emails": "thanhpk@live.<script>com"})
	fmt.Println("OUT", out)
}

func TestCompileText(t *testing.T) {
	body := `<p dir=\"ltr\"><span>aaa</span></p>`
	out := CompileTemplateToPlainText(body)
	fmt.Println("OUT", out)
}
