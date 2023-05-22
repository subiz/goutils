package template

import (
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
	body := `<p>xin chào bạn. Dưới đây là một số điểm bạn cần lưu ý</p><ul><li>Đi làm đúng giờ theo quy định ở <a href="https://subiz.com.vn">đây</a></li>
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
</script>`

	out := CompileTemplateToPlainText(body)
	fmt.Println("OUT", out)
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
