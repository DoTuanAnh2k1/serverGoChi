"""Generate cli-config reference document (.docx).

Run: python3 docs/build_cli_config_doc.py
Output: docs/cli-config-reference.docx
"""
from docx import Document
from docx.shared import Pt, RGBColor, Inches
from docx.enum.text import WD_ALIGN_PARAGRAPH
from docx.oxml.ns import qn
from docx.oxml import OxmlElement


def set_style(doc):
    normal = doc.styles["Normal"]
    normal.font.name = "Calibri"
    normal.font.size = Pt(11)


def add_code(doc, text):
    p = doc.add_paragraph()
    run = p.add_run(text)
    run.font.name = "Consolas"
    run.font.size = Pt(10)
    pf = p.paragraph_format
    pf.left_indent = Inches(0.25)
    pf.space_before = Pt(4)
    pf.space_after = Pt(4)
    shading = OxmlElement("w:shd")
    shading.set(qn("w:fill"), "F2F2F2")
    p._p.get_or_add_pPr().append(shading)
    return p


def add_h1(doc, text):
    p = doc.add_heading(text, level=1)
    return p


def add_h2(doc, text):
    p = doc.add_heading(text, level=2)
    return p


def add_h3(doc, text):
    p = doc.add_heading(text, level=3)
    return p


def add_para(doc, text, bold=False):
    p = doc.add_paragraph()
    r = p.add_run(text)
    r.bold = bold
    return p


def add_bullet(doc, text):
    p = doc.add_paragraph(text, style="List Bullet")
    return p


def add_table(doc, headers, rows):
    t = doc.add_table(rows=1 + len(rows), cols=len(headers))
    t.style = "Light Grid Accent 1"
    hdr = t.rows[0].cells
    for i, h in enumerate(headers):
        hdr[i].text = h
        for p in hdr[i].paragraphs:
            for r in p.runs:
                r.bold = True
    for ri, row in enumerate(rows, start=1):
        for ci, v in enumerate(row):
            t.rows[ri].cells[ci].text = v
    return t


def main():
    doc = Document()
    set_style(doc)

    # --- Title ---
    title = doc.add_heading("cli-config — Hướng dẫn đầy đủ các lệnh", level=0)
    title.alignment = WD_ALIGN_PARAGRAPH.CENTER
    sub = doc.add_paragraph()
    sub.alignment = WD_ALIGN_PARAGRAPH.CENTER
    srun = sub.add_run("Tham chiếu toàn bộ lệnh trong mode cli-config của SSH gate")
    srun.italic = True
    srun.font.size = Pt(11)

    # --- Overview ---
    add_h1(doc, "1. Tổng quan")
    add_para(doc,
             "cli-config là REPL (interactive shell) chạy sau khi SSH vào gate và chọn mode cli-config. "
             "Nó gọi tới mgt-svc qua HTTP bằng token đã lấy được từ bước auth để quản lý user, NE, group "
             "và các quan hệ giữa chúng. Tab để auto-complete; 'help' để xem trợ giúp; 'exit' hoặc 'quit' để quay về menu mode.")

    add_h2(doc, "Cú pháp chung")
    add_code(doc,
             "cli-config> <verb> <entity> [<target>] [<field> <value> ...]\n\n"
             "verb    : show | set | update | delete | map | unmap | help | exit\n"
             "entity  : user | ne | group\n"
             "target  : tên (name) hoặc id (số) của record\n"
             "field   : tên trường (có alias — xem bảng trường)\n"
             "value   : giá trị; nếu chứa khoảng trắng phải bỏ trong \"...\"")

    add_h2(doc, "Quy tắc chung")
    add_bullet(doc, "Field–value đi thành cặp, phân tách bằng khoảng trắng.")
    add_bullet(doc, "Giá trị có khoảng trắng phải đặt trong dấu ngoặc kép: full_name \"Alice Wonder\".")
    add_bullet(doc, "Mỗi lệnh có danh sách field BẮT BUỘC (ở lệnh set) và field TUỲ CHỌN.")
    add_bullet(doc, "account_type chỉ chấp nhận 1 (Admin) hoặc 2 (Normal). SuperAdmin không tạo được qua CLI.")
    add_bullet(doc, "conf_mode của NE chỉ chấp nhận: SSH, TELNET, NETCONF, RESTCONF.")
    add_bullet(doc, "Tab để auto-complete verb, entity, tên field và giá trị enum.")
    add_bullet(doc, "Append '--help' (hoặc '-h') vào BẤT KỲ vị trí trên dòng lệnh để in help context-aware "
                    "(ví dụ: 'set user --help' in help riêng cho 'set user'; 'show ne name HTSMF01 --help' in help cho 'show ne').")
    add_bullet(doc, "Lệnh thành công in 'OK: <mô tả>'. Lệnh lỗi in 'error: <chi tiết>' và giữ prompt.")

    # --- Field reference ---
    add_h1(doc, "2. Bảng trường cho từng entity")

    add_h2(doc, "2.1. user (tbl_account)")
    add_para(doc, "Field bắt buộc khi set (tạo mới):", bold=True)
    add_bullet(doc, "name (alias: account_name, username) — tên đăng nhập, duy nhất.")
    add_bullet(doc, "password — mật khẩu thô; mgt-svc sẽ hash trước khi lưu.")
    add_para(doc, "Field tuỳ chọn:", bold=True)
    add_table(doc,
              ["Alias", "Canonical", "Kiểu", "Ghi chú"],
              [
                  ["email", "email", "string", "Email liên hệ"],
                  ["full_name", "full_name", "string", "Họ tên đầy đủ (có thể có khoảng trắng → quote)"],
                  ["phone / phone_number", "phone_number", "string", "Số điện thoại"],
                  ["address", "address", "string", "Địa chỉ"],
                  ["description", "description", "string", "Ghi chú tự do"],
                  ["type / account_type", "account_type", "int", "1=Admin, 2=Normal"],
              ])

    add_h2(doc, "2.2. ne (cli_ne)")
    add_para(doc, "Field bắt buộc khi set (tạo mới):", bold=True)
    add_bullet(doc, "ne_name (alias: name) — tên NE.")
    add_bullet(doc, "namespace — namespace; (ne_name + namespace) duy nhất trong toàn hệ thống.")
    add_bullet(doc, "conf_master_ip (alias: ip) — IP master của NE.")
    add_bullet(doc, "conf_port_master_tcp (alias: port) — cổng TCP master.")
    add_bullet(doc, "command_url — URL/endpoint để dispatch lệnh tới NE.")
    add_para(doc, "Field tuỳ chọn:", bold=True)
    add_table(doc,
              ["Alias", "Canonical", "Kiểu", "Ghi chú"],
              [
                  ["site / site_name", "site_name", "string", "Tên site / khu vực"],
                  ["system_type", "system_type", "string", "Loại hệ thống (tự do)"],
                  ["description", "description", "string", "Ghi chú"],
                  ["mode / conf_mode", "conf_mode", "enum", "SSH | TELNET | NETCONF | RESTCONF"],
                  ["conf_slave_ip", "conf_slave_ip", "string", "IP slave (HA)"],
                  ["conf_port_master_ssh", "conf_port_master_ssh", "int", "Cổng SSH master"],
                  ["conf_port_slave_ssh", "conf_port_slave_ssh", "int", "Cổng SSH slave"],
                  ["conf_port_slave_tcp", "conf_port_slave_tcp", "int", "Cổng TCP slave"],
                  ["conf_username", "conf_username", "string", "Username để SSH vào NE"],
                  ["conf_password", "conf_password", "string", "Password để SSH vào NE"],
              ])

    add_h2(doc, "2.3. group (cli_group)")
    add_para(doc, "Field bắt buộc khi set (tạo mới):", bold=True)
    add_bullet(doc, "name — tên group, duy nhất toàn hệ thống.")
    add_para(doc, "Field tuỳ chọn:", bold=True)
    add_bullet(doc, "description — ghi chú group.")

    # --- Verbs ---
    add_h1(doc, "3. Các lệnh chi tiết")

    # ---- SHOW ----
    add_h2(doc, "3.1. show — liệt kê / xem chi tiết / lọc theo trường")
    add_code(doc,
             "show user|ne|group                      # liệt kê tất cả (bảng)\n"
             "show user|ne|group <name|id>            # legacy: tìm theo name hoặc id\n"
             "show user|ne|group <field> <value>      # lọc theo trường")

    add_para(doc, "Mandatory: không có ngoài entity. Các trường filter hỗ trợ:", bold=True)
    add_table(doc,
              ["Entity", "Filter field (alias)", "Hành vi"],
              [
                  ["user",
                   "name (username, account_name) | id (account_id) | email | role (type, account_type)",
                   "name/id/email: match chính xác, in detail (table nếu trùng). role: luôn in bảng (nhiều user có thể cùng role). role nhận label SuperAdmin|Admin|Normal hoặc số 0|1|2."],
                  ["ne",
                   "name (ne_name) | id | site (site_name) | namespace",
                   "name: in bảng nếu ne_name trùng qua nhiều namespace; id: in detail. site/namespace: luôn in bảng."],
                  ["group",
                   "name | id",
                   "In detail (kèm users và ne_ids)."],
              ])
    add_para(doc, "Tab ở vị trí field sẽ gợi ý các alias; Tab ở vị trí value của trường enum (role) sẽ gợi ý SuperAdmin/Admin/Normal.")

    add_h3(doc, "Ví dụ & output — user")
    add_code(doc, "cli-config> show user")
    add_code(doc,
             "ID  NAME       ROLE        ENABLED  EMAIL              FULL NAME\n"
             "1   anhdt195   SuperAdmin  true                         \n"
             "2   alice      Admin       true     alice@example.com   Alice Wonder\n"
             "3   bob        Normal      true     bob@example.com     Bob B\n"
             "(3 users)")

    add_code(doc, "cli-config> show user name alice")
    add_code(doc,
             "id:           2\n"
             "name:         alice\n"
             "role:         Admin\n"
             "enabled:      true\n"
             "email:        alice@example.com\n"
             "full_name:    Alice Wonder\n"
             "phone:        0900000000\n"
             "address:      HN\n"
             "description:  demo user\n"
             "created_by:   anhdt195")

    add_code(doc, "cli-config> show user alice              # legacy: cùng nghĩa 'show user name alice'")
    add_code(doc, "cli-config> show user email bob@example.com")
    add_code(doc,
             "id:           3\n"
             "name:         bob\n"
             "role:         Normal\n"
             "...")

    add_code(doc, "cli-config> show user role Admin         # in toàn bộ Admin")
    add_code(doc,
             "ID  NAME     ROLE   ENABLED  EMAIL              FULL NAME\n"
             "2   alice    Admin  true     alice@example.com  Alice Wonder\n"
             "4   carol    Admin  true     carol@example.com  Carol C\n"
             "(2 users)")

    add_code(doc, "cli-config> show user role 2             # giá trị số: Normal")
    add_code(doc,
             "ID  NAME  ROLE    ENABLED  EMAIL            FULL NAME\n"
             "3   bob   Normal  true     bob@example.com  Bob B\n"
             "(1 user)")

    add_h3(doc, "Ví dụ & output — ne")
    add_code(doc, "cli-config> show ne")
    add_code(doc,
             "ID  NAME     SITE  NAMESPACE  IP            PORT  MODE\n"
             "1   HTSMF01  HN    default    10.0.0.10     830   NETCONF\n"
             "2   HTSMF01  HCM   tenant-a   10.1.0.10     830   NETCONF\n"
             "3   HTSMF02  HN    default    10.0.0.11     22    SSH\n"
             "(3 NEs)")

    add_code(doc, "cli-config> show ne name HTSMF01         # trùng tên qua nhiều namespace → bảng")
    add_code(doc,
             "ID  NAME     SITE  NAMESPACE  IP          PORT  MODE\n"
             "1   HTSMF01  HN    default    10.0.0.10   830   NETCONF\n"
             "2   HTSMF01  HCM   tenant-a   10.1.0.10   830   NETCONF\n"
             "(2 NEs)")

    add_code(doc, "cli-config> show ne id 3                 # id luôn duy nhất → detail")
    add_code(doc,
             "id:                    3\n"
             "ne_name:               HTSMF02\n"
             "site_name:             HN\n"
             "namespace:             default\n"
             "conf_master_ip:        10.0.0.11\n"
             "conf_port_master_tcp:  22\n"
             "command_url:           ssh://10.0.0.11\n"
             "conf_mode:             SSH")

    add_code(doc, "cli-config> show ne site HN              # tất cả NE ở site HN")
    add_code(doc,
             "ID  NAME     SITE  NAMESPACE  IP          PORT  MODE\n"
             "1   HTSMF01  HN    default    10.0.0.10   830   NETCONF\n"
             "3   HTSMF02  HN    default    10.0.0.11   22    SSH\n"
             "(2 NEs)")

    add_code(doc, "cli-config> show ne namespace tenant-a   # tất cả NE trong namespace")
    add_code(doc,
             "ID  NAME     SITE  NAMESPACE  IP          PORT  MODE\n"
             "2   HTSMF01  HCM   tenant-a   10.1.0.10   830   NETCONF\n"
             "(1 NE)")

    add_h3(doc, "Ví dụ & output — group")
    add_code(doc, "cli-config> show group")
    add_code(doc,
             "ID  NAME  DESCRIPTION\n"
             "3   dev   dev team\n"
             "(1 group)")

    add_code(doc, "cli-config> show group name dev          # detail")
    add_code(doc,
             "id:           3\n"
             "name:         dev\n"
             "description:  dev team\n"
             "users:        alice, bob\n"
             "ne_ids:       1, 2")

    add_code(doc, "cli-config> show group dev               # legacy: cùng nghĩa 'show group name dev'")

    # ---- SET ----
    add_h2(doc, "3.2. set — tạo mới")
    add_code(doc, "set user|ne|group <field> <value> [<field> <value> ...]")
    add_para(doc,
             "Tất cả field bắt buộc của entity phải có mặt; thiếu field bắt buộc sẽ trả lỗi và huỷ lệnh.")

    add_h3(doc, "Ví dụ — user")
    add_code(doc,
             "cli-config> set user name alice password secret email alice@example.com \\\n"
             "            full_name \"Alice Wonder\" phone 0900000000 type 2")
    add_code(doc, "OK: user created")
    add_para(doc, "Tối thiểu (chỉ field bắt buộc):")
    add_code(doc, "cli-config> set user name bob password 123456")
    add_code(doc, "OK: user created")

    add_para(doc, "Re-enable user đã bị disable (merge fields mới):", bold=True)
    add_bullet(doc,
               "Nếu account_name đã tồn tại nhưng đang disable → server merge các field "
               "non-empty từ request vào record cũ, bật is_enable=true, trả 201.")
    add_bullet(doc, "Field không truyền trong request giữ nguyên giá trị cũ.")
    add_bullet(doc, "Password luôn được refresh theo giá trị request.")
    add_bullet(doc,
               "Email: EnsureEmailUnique bỏ qua tài khoản disable, nghĩa là email "
               "thuộc user đã disable coi như free — có thể tái sử dụng cho user mới "
               "hoặc tiếp tục dùng khi re-enable chính user đó.")
    add_code(doc,
             "# alice đang bị disable, email cũ 'old@example.com', phone 0900000000\n"
             "cli-config> set user name alice password newpass email new@example.com \\\n"
             "            full_name \"Alice New\"\n"
             "OK: user created        # 201, re-enabled với email/full_name mới;\n"
             "                        # phone giữ 0900000000 vì request không truyền")

    add_h3(doc, "Ví dụ — ne")
    add_code(doc,
             "cli-config> set ne ne_name HTSMF01 namespace default ip 10.0.0.10 \\\n"
             "            port 830 command_url http://10.0.0.10/restconf \\\n"
             "            mode NETCONF site_name HN system_type smf \\\n"
             "            conf_username admin conf_password netadmin \\\n"
             "            description \"primary SMF\"")
    add_code(doc, "OK: NE created")
    add_para(doc, "Tối thiểu:")
    add_code(doc,
             "cli-config> set ne ne_name HTSMF02 namespace default ip 10.0.0.11 \\\n"
             "            port 22 command_url ssh://10.0.0.11")
    add_code(doc, "OK: NE created")

    add_h3(doc, "Ví dụ — group")
    add_code(doc, "cli-config> set group name dev description \"dev team\"")
    add_code(doc, "OK: group created")

    # ---- UPDATE ----
    add_h2(doc, "3.3. update — sửa field")
    add_code(doc, "update user|ne|group <name|id> <field> <value> [<field> <value> ...]")
    add_para(doc,
             "Chỉ cần truyền field muốn đổi. Target là tên (hoặc id với ne / group). "
             "Với user, target là account_name.")

    add_h3(doc, "Ví dụ")
    add_code(doc, "cli-config> update user alice email new@example.com phone 0911111111")
    add_code(doc, "OK: user updated")
    add_code(doc, "cli-config> update ne HTSMF01 site_name HN2 description \"moved site\"")
    add_code(doc, "OK: NE updated")
    add_code(doc, "cli-config> update group 3 description \"dev team v2\"")
    add_code(doc, "OK: group updated")

    # ---- DELETE ----
    add_h2(doc, "3.4. delete — xoá record (có prompt xác nhận)")
    add_code(doc, "delete user|ne|group <name|id>")
    add_bullet(doc, "user: target là account_name. SuperAdmin KHÔNG thể bị xoá.")
    add_bullet(doc, "ne:   target là ne_name hoặc id; xoá cascade mapping user↔NE, group↔NE.")
    add_bullet(doc, "group: target là name hoặc id.")
    add_para(doc, "Mọi lệnh delete đều hỏi confirm trước khi thực hiện:", bold=True)
    add_bullet(doc, "Nhập 'y' hoặc 'yes' (không phân biệt hoa thường) → xoá.")
    add_bullet(doc, "Bỏ trống + Enter, 'n', hoặc bất kỳ input khác → abort, in 'aborted'.")
    add_bullet(doc, "Ctrl+C / Ctrl+D / đóng session → abort.")

    add_h3(doc, "Ví dụ — confirm yes")
    add_code(doc,
             "cli-config> delete user alice\n"
             "Delete user \"alice\"? [y/N]: y\n"
             "OK: user deleted")

    add_h3(doc, "Ví dụ — confirm no / abort")
    add_code(doc,
             "cli-config> delete ne HTSMF02\n"
             "Delete NE \"HTSMF02\"? [y/N]: \n"
             "aborted")
    add_code(doc,
             "cli-config> delete group dev\n"
             "Delete group \"dev\"? [y/N]: n\n"
             "aborted")

    # ---- MAP / UNMAP ----
    add_h2(doc, "3.5. map / unmap — quan hệ")
    add_code(doc,
             "map user <username> ne <ne_name|id>\n"
             "map user <username> group <group_name|id>\n"
             "map group <group_name|id> ne <ne_name|id>\n"
             "unmap ...  (cùng shape với map)")
    add_para(doc,
             "User có quyền lên NE theo 2 cách: (a) map user trực tiếp tới NE, hoặc (b) user thuộc group, "
             "group được map với NE. Unmap chỉ gỡ đúng quan hệ được chỉ định; nếu user còn quyền qua group thì "
             "vẫn tiếp cận được NE.")

    add_h3(doc, "Ví dụ")
    add_code(doc,
             "cli-config> map user alice ne HTSMF01\n"
             "OK: map user↔NE\n"
             "cli-config> map user alice group dev\n"
             "OK: map user↔group\n"
             "cli-config> map group dev ne HTSMF02\n"
             "OK: map group↔NE")

    add_code(doc,
             "cli-config> unmap user alice ne HTSMF01\n"
             "OK: unmap user↔NE\n"
             "cli-config> unmap group dev ne HTSMF02\n"
             "OK: unmap group↔NE")

    add_para(doc, "Không hỗ trợ:")
    add_bullet(doc, "map ne ... — không có chiều này, dùng map user hoặc map group.")

    # ---- HELP / EXIT ----
    add_h2(doc, "3.6. help / --help")
    add_code(doc,
             "help                          # trợ giúp tổng quan\n"
             "help <verb>                   # chi tiết 1 verb (show, set, update, delete, map, exit, help)\n"
             "help <verb> <entity>          # chi tiết chuyên biệt cho verb+entity (ví dụ 'help set user')\n"
             "<bất kỳ lệnh nào> --help      # same (context-aware)\n"
             "<bất kỳ lệnh nào> -h          # alias ngắn của --help")
    add_para(doc, "Help context-aware: nếu có entity ngay sau verb, help topic sẽ là 'verb entity' (ví dụ 'set user'); "
                  "fallback về verb-only nếu không có entry chuyên biệt. Ví dụ:")
    add_code(doc,
             "cli-config> set user --help\n"
             "set user name <name> password <password> [<field> <value> ...]\n"
             "\n"
             "Required fields:\n"
             "  name (alias: account_name, username)\n"
             "  password\n"
             "\n"
             "Optional fields:\n"
             "  email, full_name, phone, address, description,\n"
             "  type (1=Admin, 2=Normal)\n"
             "\n"
             "Example:\n"
             "  set user name alice password secret email alice@example.com \\\n"
             "      full_name \"Alice Wonder\" phone 0900000000 type 2")

    add_h2(doc, "3.7. exit / quit")
    add_para(doc, "Thoát mode cli-config, quay lại menu 'mode>'. Không nhận tham số.")

    add_h2(doc, "3.8. RBAC — NE profile, command registry, group permission")
    add_para(doc,
             "Hệ thống RBAC (docs/rbac-design.md) bổ sung 3 entity mới và các verb "
             "allow / deny / revoke. Dùng để trả lời câu hỏi \"user X được chạy command Z "
             "trên NE Y không?\" cho downstream ne-config / ne-command.")
    add_para(doc, "Entity mới:", bold=True)
    add_bullet(doc, "ne-profile — phân loại NE theo tập lệnh (SMF / AMF / UPF / generic-router ...).")
    add_bullet(doc, "command-def — registry lệnh; mỗi def có (service, ne_profile, pattern, category, risk_level).")
    add_bullet(doc, "command-group — bundle nhiều command-def cùng ne_profile, để gán permission 1 lần.")
    add_para(doc, "Verb mới:", bold=True)
    add_bullet(doc, "allow <group> <grant_type> <grant_value> [ne_scope <scope>] [service <svc>]")
    add_bullet(doc, "deny  <group> <grant_type> <grant_value> [ne_scope <scope>] [service <svc>]")
    add_bullet(doc, "revoke <group> <perm_id>  (xoá 1 rule theo id).")
    add_para(doc, "grant_type ∈ {command-group, category, pattern}. ne_scope ∈ {\"*\", \"profile:<name>\", \"ne:<ne_name>\"}.")
    add_para(doc,
             "Evaluation: kết hợp AWS-IAM (explicit deny > explicit allow > implicit deny) "
             "với scope specificity (ne > profile > *). Tại scope cụ thể nhất có match, deny thắng allow.")
    add_h3(doc, "Kịch bản ví dụ (end-to-end)")
    add_code(doc,
             "# 1) Tạo profile và gán cho NE\n"
             "cli-config> set ne-profile name SMF description \"Session Management Function\"\n"
             "OK: ne-profile created\n"
             "cli-config> update ne HTSMF01 ne_profile SMF\n"
             "OK: NE updated\n"
             "\n"
             "# 2) Khai báo command-def\n"
             "cli-config> set command-def service ne-command ne_profile SMF \\\n"
             "            pattern \"get subscriber\" category monitoring\n"
             "OK: command-def created\n"
             "cli-config> set command-def service ne-command ne_profile SMF \\\n"
             "            pattern \"get session\" category monitoring\n"
             "OK: command-def created\n"
             "\n"
             "# 3) Gom thành group\n"
             "cli-config> set command-group name smf-subscriber-ops ne_profile SMF service ne-command\n"
             "OK: command-group created\n"
             "cli-config> map command-group smf-subscriber-ops command 1\n"
             "OK: map command-group↔command\n"
             "\n"
             "# 4) Gán permission cho user group\n"
             "cli-config> allow team-smf-l1 command-group smf-subscriber-ops ne_scope profile:SMF\n"
             "OK: allow command_group=smf-subscriber-ops on scope=profile:SMF added to group \"team-smf-l1\"\n"
             "cli-config> deny team-smf-l1 pattern \"delete *\" ne_scope ne:SMF-01\n"
             "OK: deny pattern=delete * on scope=ne:SMF-01 added to group \"team-smf-l1\"\n"
             "\n"
             "# 5) Kiểm tra / gỡ\n"
             "cli-config> show command-group smf-subscriber-ops\n"
             "id:          1  name: smf-subscriber-ops  ...\n"
             "members:     1:get subscriber\n"
             "cli-config> revoke team-smf-l1 42\n"
             "Delete permission \"42\"? [y/N]: y\n"
             "OK: permission revoked")

    add_h2(doc, "3.9. Giới hạn mode khi SSH")
    add_para(doc,
             "Sau khi SSH login thành công: SuperAdmin / Admin thấy đủ 3 mode; Normal user "
             "chỉ thấy ne-config và ne-command — cli-config bị ẩn khỏi menu 'mode>'. "
             "Whitelist command cho Normal user được 2 service downstream tự fetch từ "
             "mgt-svc qua endpoint /aa/authorize/rbac/effective (cache mỗi session) hoặc "
             "/aa/authorize/rbac/check-command (realtime).")

    # --- Session walkthrough ---
    add_h1(doc, "4. Kịch bản mẫu end-to-end")
    add_para(doc,
             "Ví dụ đầy đủ từ lúc SSH vào gate đến khi tạo user, NE, group và map quyền cho user.")
    add_code(doc,
             "$ ssh anhdt195@gate.example.com -p 2223\n"
             "anhdt195@gate.example.com's password: ***\n"
             "\n"
             "Welcome anhdt195 — management CLI.\n"
             "Available modes: cli-config, ne-config, ne-command (Tab to cycle / autocomplete, 'exit' to quit).\n"
             "\n"
             "mode> cli-config\n"
             "\n"
             "== cli-config mode ==\n"
             "Type 'help' for commands. Type 'exit' to return to the mode menu.\n"
             "\n"
             "cli-config> set user name alice password secret type 2 email alice@example.com\n"
             "OK: user created\n"
             "cli-config> set group name dev description \"dev team\"\n"
             "OK: group created\n"
             "cli-config> set ne ne_name HTSMF01 namespace default ip 10.0.0.10 port 830 \\\n"
             "            command_url http://10.0.0.10/restconf mode NETCONF site_name HN\n"
             "OK: NE created\n"
             "cli-config> map user alice group dev\n"
             "OK: map user↔group\n"
             "cli-config> map group dev ne HTSMF01\n"
             "OK: map group↔NE\n"
             "cli-config> show user alice\n"
             "id:           2\n"
             "name:         alice\n"
             "role:         Normal\n"
             "enabled:      true\n"
             "email:        alice@example.com\n"
             "full_name:    \n"
             "phone:        \n"
             "address:      \n"
             "description:  \n"
             "created_by:   anhdt195\n"
             "cli-config> show group dev\n"
             "id:           3\n"
             "name:         dev\n"
             "description:  dev team\n"
             "users:        alice\n"
             "ne_ids:       1\n"
             "cli-config> exit\n"
             "mode> exit\n"
             "bye.\n")

    # --- Errors ---
    add_h1(doc, "5. Lỗi thường gặp")
    add_table(doc,
              ["Thông báo", "Nguyên nhân", "Cách xử lý"],
              [
                  ["unknown command \"x\"",
                   "Verb không phải show/set/update/delete/map/unmap/help/exit.",
                   "Dùng Tab hoặc gõ 'help' để xem verb hợp lệ."],
                  ["unknown field \"x\" for user (valid: ...)",
                   "Alias trường không hợp lệ.",
                   "Xem bảng trường, dùng Tab để hiện gợi ý."],
                  ["missing required field \"X\" for set ne (required: ...)",
                   "Thiếu field bắt buộc khi set.",
                   "Bổ sung đủ field; tham chiếu bảng ở mục 2."],
                  ["field \"conf_mode\" must be one of [SSH TELNET NETCONF RESTCONF], got \"...\"",
                   "Giá trị enum không hợp lệ.",
                   "Đặt đúng một trong các giá trị cho phép."],
                  ["field \"port\" must be an integer, got \"abc\"",
                   "Field kiểu số truyền vào chuỗi.",
                   "Truyền số nguyên, vd 830."],
                  ["no user with name or id \"x\"",
                   "Không tìm thấy target khi show/update/delete.",
                   "Dùng 'show user' để kiểm tra tên/id hiện có."],
                  ["unterminated quoted string",
                   "Dấu nháy kép mở mà chưa đóng.",
                   "Đóng lại \" hoặc bỏ quote."],
              ])

    # --- Tips ---
    add_h1(doc, "6. Mẹo sử dụng")
    add_bullet(doc, "Tab cycling: nhấn Tab liên tục để quay vòng qua các candidate ở vị trí con trỏ.")
    add_bullet(doc, "Hint hiện ở dòng ngay dưới prompt khi có từ 2 candidate trở lên.")
    add_bullet(doc, "Alias: dùng 'name', 'ip', 'port', 'mode', 'type'… thay cho tên canonical để gõ nhanh.")
    add_bullet(doc, "Lệnh có thể dài vô hạn; terminal tự word-wrap theo kích thước PTY thật của client.")
    add_bullet(doc, "Xác thực: username + password user khi SSH chính là tài khoản mgt-svc; role quyết định CRUD được phép.")

    out = "/home/phatlc/data/serverGoChi/docs/cli-config-reference.docx"
    doc.save(out)
    print(out)


if __name__ == "__main__":
    main()
